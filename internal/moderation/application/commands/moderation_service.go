package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	communitydomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/domain"
	dictdomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/domain"
	identitydomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/identity/domain"
	moderationdomain "github.com/iqbaleff214/kamus-banjar-api-2/internal/moderation/domain"
)

type ModerationService struct {
	modRepo  ModerationRepository
	contribs ContributionWriter
	words    WordWriter
	users    UserWriter
}

func NewModerationService(
	modRepo ModerationRepository,
	contribs ContributionWriter,
	words WordWriter,
	users UserWriter,
) *ModerationService {
	return &ModerationService{
		modRepo:  modRepo,
		contribs: contribs,
		words:    words,
		users:    users,
	}
}

// ─── Queues / stats ───────────────────────────────────────────────────────────

func (s *ModerationService) GetModerationQueue(ctx context.Context, ctype *communitydomain.ContributionType, page, perPage int) ([]*communitydomain.Contribution, int, error) {
	return s.modRepo.GetPendingContributions(ctx, ctype, page, perPage)
}

func (s *ModerationService) GetFlaggedComments(ctx context.Context, page, perPage int) ([]*communitydomain.Comment, int, error) {
	return s.modRepo.GetFlaggedComments(ctx, page, perPage)
}

func (s *ModerationService) GetStats(ctx context.Context) (*moderationdomain.ModerationStats, error) {
	return s.modRepo.GetStats(ctx)
}

func (s *ModerationService) ListUsers(ctx context.Context, role, query string, isActive *bool, page, perPage int) ([]*identitydomain.User, int, error) {
	return s.modRepo.ListUsers(ctx, role, query, isActive, page, perPage)
}

func (s *ModerationService) GetUser(ctx context.Context, userID uuid.UUID) (*identitydomain.User, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

// ─── Contribution moderation ──────────────────────────────────────────────────

func (s *ModerationService) ApproveContribution(ctx context.Context, adminID, contributionID uuid.UUID, note string) error {
	c, err := s.contribs.FindByID(ctx, contributionID)
	if err != nil {
		return ErrContributionNotFound
	}

	if err := c.Approve(adminID, note); err != nil {
		return ErrContributionConflict
	}

	if err := s.mergeContribution(ctx, c); err != nil {
		return err
	}

	if err := s.contribs.Update(ctx, c); err != nil {
		return err
	}

	log, _ := moderationdomain.NewAuditLog(adminID, moderationdomain.ActionApproveContribution, "contribution", contributionID, nil)
	return s.modRepo.CreateAuditLog(ctx, log)
}

func (s *ModerationService) RejectContribution(ctx context.Context, adminID, contributionID uuid.UUID, note string) error {
	if note == "" {
		return ErrNoteRequired
	}

	c, err := s.contribs.FindByID(ctx, contributionID)
	if err != nil {
		return ErrContributionNotFound
	}

	if err := c.Reject(adminID, note); err != nil {
		return ErrContributionConflict
	}

	if err := s.contribs.Update(ctx, c); err != nil {
		return err
	}

	log, _ := moderationdomain.NewAuditLog(adminID, moderationdomain.ActionRejectContribution, "contribution", contributionID, map[string]any{"note": note})
	return s.modRepo.CreateAuditLog(ctx, log)
}

// mergeContribution applies the approved contribution's payload to the dictionary.
func (s *ModerationService) mergeContribution(ctx context.Context, c *communitydomain.Contribution) error {
	switch c.Type {
	case communitydomain.ContributionTypeNewWord:
		return s.mergeNewWord(ctx, c)
	case communitydomain.ContributionTypeNewDefinition:
		return s.mergeNewDefinition(ctx, c)
	case communitydomain.ContributionTypeNewExample:
		return s.mergeNewExample(ctx, c)
	case communitydomain.ContributionTypeEditWord:
		return s.mergeEditWord(ctx, c)
	}
	return nil
}

func (s *ModerationService) mergeNewWord(ctx context.Context, c *communitydomain.Contribution) error {
	banjar, _ := c.Payload["banjar"].(string)
	wordClass, _ := c.Payload["word_class"].(string)
	if banjar == "" || wordClass == "" {
		return ErrInvalidPayload
	}

	word, err := dictdomain.NewWord(banjar, dictdomain.WordClass(wordClass), dictdomain.DialectHulu)
	if err != nil {
		return fmt.Errorf("new word: %w", err)
	}
	word.Source = dictdomain.SourceContributed
	word.CreatedBy = &c.ContributorID

	if syl, ok := c.Payload["banjar_syllabified"].(string); ok {
		word.BanjarSyllabified = syl
	}

	if defs, ok := c.Payload["definitions"].([]any); ok {
		for i, d := range defs {
			dm, _ := d.(map[string]any)
			meaning, _ := dm["meaning"].(string)
			if meaning != "" {
				_ = word.AddDefinition(meaning, i+1)
			}
		}
	}

	if exs, ok := c.Payload["examples"].([]any); ok {
		for _, e := range exs {
			em, _ := e.(map[string]any)
			banjarSent, _ := em["banjar_sentence"].(string)
			indonesian, _ := em["indonesian_translation"].(string)
			if banjarSent != "" {
				word.AddExample(banjarSent, indonesian)
			}
		}
	}

	return s.words.Create(ctx, word)
}

func (s *ModerationService) mergeNewDefinition(ctx context.Context, c *communitydomain.Contribution) error {
	if c.TargetWordID == nil {
		return ErrInvalidPayload
	}
	meaning, _ := c.Payload["meaning"].(string)
	if meaning == "" {
		return ErrInvalidPayload
	}

	word, err := s.words.FindByID(ctx, *c.TargetWordID)
	if err != nil {
		return ErrWordNotFound
	}

	if err := word.AddDefinition(meaning, len(word.Definitions)+1); err != nil {
		return err
	}
	return s.words.Update(ctx, word)
}

func (s *ModerationService) mergeNewExample(ctx context.Context, c *communitydomain.Contribution) error {
	if c.TargetWordID == nil {
		return ErrInvalidPayload
	}
	banjarSent, _ := c.Payload["banjar_sentence"].(string)
	indonesian, _ := c.Payload["indonesian_translation"].(string)
	if banjarSent == "" {
		return ErrInvalidPayload
	}

	word, err := s.words.FindByID(ctx, *c.TargetWordID)
	if err != nil {
		return ErrWordNotFound
	}

	word.AddExample(banjarSent, indonesian)
	return s.words.Update(ctx, word)
}

func (s *ModerationService) mergeEditWord(ctx context.Context, c *communitydomain.Contribution) error {
	if c.TargetWordID == nil {
		return ErrInvalidPayload
	}

	word, err := s.words.FindByID(ctx, *c.TargetWordID)
	if err != nil {
		return ErrWordNotFound
	}

	if banjar, ok := c.Payload["banjar"].(string); ok && banjar != "" {
		word.Banjar = banjar
	}
	if syl, ok := c.Payload["banjar_syllabified"].(string); ok {
		word.BanjarSyllabified = syl
	}
	if wc, ok := c.Payload["word_class"].(string); ok && wc != "" {
		if err := dictdomain.WordClass(wc).Validate(); err != nil {
			return err
		}
		word.WordClass = dictdomain.WordClass(wc)
	}
	return s.words.Update(ctx, word)
}

// ─── User moderation ──────────────────────────────────────────────────────────

func (s *ModerationService) BanUser(ctx context.Context, adminID, targetUserID uuid.UUID, reason string) (*identitydomain.User, error) {
	target, err := s.users.FindByID(ctx, targetUserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if target.Role == identitydomain.RoleAdmin {
		return nil, moderationdomain.ErrCannotBanAdmin
	}

	target.Ban()

	if err := s.users.Update(ctx, target); err != nil {
		return nil, err
	}

	meta := map[string]any{"reason": reason}
	log, _ := moderationdomain.NewAuditLog(adminID, moderationdomain.ActionBanUser, "user", targetUserID, meta)
	return target, s.modRepo.CreateAuditLog(ctx, log)
}

func (s *ModerationService) UnbanUser(ctx context.Context, adminID, targetUserID uuid.UUID) (*identitydomain.User, error) {
	target, err := s.users.FindByID(ctx, targetUserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	target.Unban()

	if err := s.users.Update(ctx, target); err != nil {
		return nil, err
	}

	log, _ := moderationdomain.NewAuditLog(adminID, moderationdomain.ActionUnbanUser, "user", targetUserID, nil)
	return target, s.modRepo.CreateAuditLog(ctx, log)
}

func (s *ModerationService) ChangeUserRole(ctx context.Context, adminID, targetUserID uuid.UUID, role identitydomain.Role) (*identitydomain.User, error) {
	if role != identitydomain.RoleUser && role != identitydomain.RoleAdmin {
		return nil, ErrInvalidRole
	}

	if adminID == targetUserID {
		return nil, moderationdomain.ErrCannotSelfDemote
	}

	target, err := s.users.FindByID(ctx, targetUserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if err := target.Promote(role); err != nil {
		return nil, err
	}

	if err := s.users.Update(ctx, target); err != nil {
		return nil, err
	}

	meta := map[string]any{"new_role": string(role)}
	log, _ := moderationdomain.NewAuditLog(adminID, moderationdomain.ActionChangeRole, "user", targetUserID, meta)
	return target, s.modRepo.CreateAuditLog(ctx, log)
}
