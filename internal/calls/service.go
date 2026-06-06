package calls

import (
	"context"
	"errors"
	"time"

	db "github.com/brainart16/brenox/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service struct {
	queries    *db.Queries
	broadcast  Broadcaster
	invites    InviteNotifier
	access     ChannelAccessChecker
}

func NewService(
	queries *db.Queries,
	broadcast Broadcaster,
	invites InviteNotifier,
	access ChannelAccessChecker,
) *Service {
	return &Service{
		queries:   queries,
		broadcast: broadcast,
		invites:   invites,
		access:    access,
	}
}

func (s *Service) InitiateCall(
	ctx context.Context,
	workspaceID, channelID, userID int64,
) (CallResponse, error) {
	if err := s.access.AssertChannelMember(ctx, workspaceID, channelID, userID); err != nil {
		return CallResponse{}, mapAccessErr(err)
	}

	if _, err := s.queries.GetActiveCallByChannel(ctx, channelID); err == nil {
		return CallResponse{}, ErrCallAlreadyActive
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return CallResponse{}, err
	}

	call, err := s.queries.CreateCall(ctx, db.CreateCallParams{
		ChannelID:   channelID,
		WorkspaceID: workspaceID,
		InitiatorID: userID,
		Status:      StatusRinging,
	})
	if err != nil {
		return CallResponse{}, err
	}

	if _, err := s.queries.AddCallParticipant(ctx, db.AddCallParticipantParams{
		CallID: call.ID,
		UserID: userID,
	}); err != nil {
		return CallResponse{}, err
	}

	s.publish(ctx, "call.join", call, userID)
	s.notifyChannelMembers(ctx, call, userID)

	return toCallResponse(call), nil
}

func (s *Service) JoinCall(ctx context.Context, callID, userID int64) (CallResponse, error) {
	call, err := s.queries.GetCallByID(ctx, callID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CallResponse{}, ErrNotFound
		}
		return CallResponse{}, err
	}
	if call.Status == StatusEnded {
		return CallResponse{}, ErrCallEnded
	}

	if err := s.access.AssertChannelMember(ctx, call.WorkspaceID, call.ChannelID, userID); err != nil {
		return CallResponse{}, mapAccessErr(err)
	}

	if _, err := s.queries.GetActiveCallParticipant(ctx, db.GetActiveCallParticipantParams{
		CallID: callID,
		UserID: userID,
	}); errors.Is(err, pgx.ErrNoRows) {
		if _, err := s.queries.AddCallParticipant(ctx, db.AddCallParticipantParams{
			CallID: callID,
			UserID: userID,
		}); err != nil {
			return CallResponse{}, err
		}
	} else if err != nil {
		return CallResponse{}, err
	}

	count, err := s.queries.CountActiveCallParticipants(ctx, callID)
	if err != nil {
		return CallResponse{}, err
	}

	if call.Status == StatusRinging && count >= 2 {
		call, err = s.queries.UpdateCallStatus(ctx, db.UpdateCallStatusParams{
			ID:     callID,
			Status: StatusActive,
		})
		if err != nil {
			return CallResponse{}, err
		}
	}

	s.publish(ctx, "call.join", call, userID)
	return toCallResponse(call), nil
}

func (s *Service) LeaveCall(ctx context.Context, callID, userID int64) (CallResponse, error) {
	call, err := s.queries.GetCallByID(ctx, callID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CallResponse{}, ErrNotFound
		}
		return CallResponse{}, err
	}
	if call.Status == StatusEnded {
		return toCallResponse(call), nil
	}

	if _, err := s.queries.GetActiveCallParticipant(ctx, db.GetActiveCallParticipantParams{
		CallID: callID,
		UserID: userID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CallResponse{}, ErrNotParticipant
		}
		return CallResponse{}, err
	}

	if _, err := s.queries.MarkCallParticipantLeft(ctx, db.MarkCallParticipantLeftParams{
		CallID: callID,
		UserID: userID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CallResponse{}, ErrNotParticipant
		}
		return CallResponse{}, err
	}

	s.publish(ctx, "call.leave", call, userID)

	count, err := s.queries.CountActiveCallParticipants(ctx, callID)
	if err != nil {
		return CallResponse{}, err
	}

	if count == 0 {
		call, err = s.queries.UpdateCallStatus(ctx, db.UpdateCallStatusParams{
			ID:     callID,
			Status: StatusEnded,
		})
		if err != nil {
			return CallResponse{}, err
		}
		s.publish(ctx, "call.end", call, userID)
	}

	return toCallResponse(call), nil
}

func (s *Service) ValidateSignal(ctx context.Context, callID, userID int64) (SignalContext, error) {
	call, err := s.queries.GetCallByID(ctx, callID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SignalContext{}, ErrNotFound
		}
		return SignalContext{}, err
	}
	if call.Status == StatusEnded {
		return SignalContext{}, ErrCallEnded
	}

	if _, err := s.queries.GetActiveCallParticipant(ctx, db.GetActiveCallParticipantParams{
		CallID: callID,
		UserID: userID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SignalContext{}, ErrNotParticipant
		}
		return SignalContext{}, err
	}

	return SignalContext{
		CallID:      call.ID,
		WorkspaceID: call.WorkspaceID,
		ChannelID:   call.ChannelID,
		UserID:      userID,
	}, nil
}

func (s *Service) publish(ctx context.Context, eventType string, call db.Call, userID int64) {
	if s.broadcast == nil {
		return
	}
	s.broadcast.PublishCallEvent(eventType, call.WorkspaceID, call.ChannelID, map[string]any{
		"call_id":      call.ID,
		"user_id":      userID,
		"channel_id":   call.ChannelID,
		"workspace_id": call.WorkspaceID,
		"status":       call.Status,
	})
}

func (s *Service) notifyChannelMembers(ctx context.Context, call db.Call, initiatorID int64) {
	if s.invites == nil {
		return
	}

	initiator, err := s.queries.GetUserByID(ctx, initiatorID)
	if err != nil {
		return
	}

	userIDs, err := s.queries.ListChannelMemberUserIDs(ctx, call.ChannelID)
	if err != nil {
		return
	}

	for _, targetID := range userIDs {
		if targetID == initiatorID {
			continue
		}
		_ = s.invites.NotifyCallInvite(ctx, call.WorkspaceID, call.ChannelID, call.ID, initiatorID, targetID, initiator.Username)
	}
}

func mapAccessErr(err error) error {
	if errors.Is(err, ErrNotMember) {
		return ErrNotMember
	}
	return err
}

func toCallResponse(call db.Call) CallResponse {
	resp := CallResponse{
		ID:          call.ID,
		ChannelID:   call.ChannelID,
		WorkspaceID: call.WorkspaceID,
		InitiatorID: call.InitiatorID,
		Status:      call.Status,
		CreatedAt:   formatTime(call.CreatedAt),
	}
	if call.EndedAt.Valid {
		resp.EndedAt = formatTime(call.EndedAt)
	}
	return resp
}

func formatTime(ts pgtype.Timestamptz) string {
	if !ts.Valid {
		return ""
	}
	return ts.Time.UTC().Format(time.RFC3339Nano)
}
