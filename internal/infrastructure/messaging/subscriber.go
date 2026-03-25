package messaging

import (
	"context"
	"encoding/json"

	appuser "github.com/chengtb/chengtb/internal/application/user"
	"github.com/chengtb/chengtb/pkg/logger"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// Subjects used for the User micro-service.
const (
	SubjectCreateUser = "user.create"
	SubjectGetUser    = "user.get"
	SubjectListUsers  = "user.list"
	SubjectUpdateUser = "user.update"
	SubjectDeleteUser = "user.delete"
)

// --- generic helpers ---------------------------------------------------------

// replyOK serialises payload as JSON and writes it to the reply subject.
func replyOK(msg *nats.Msg, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		replyErr(msg, "internal error: marshal failed")
		return
	}
	if err := msg.Respond(data); err != nil {
		logger.L().Error("nats respond error", zap.Error(err))
	}
}

// replyErr sends a JSON error envelope to the reply subject.
func replyErr(msg *nats.Msg, errMsg string) {
	type errEnvelope struct {
		Error string `json:"error"`
	}
	data, _ := json.Marshal(errEnvelope{Error: errMsg})
	if err := msg.Respond(data); err != nil {
		logger.L().Error("nats respond error", zap.Error(err))
	}
}

// --- subscriber -------------------------------------------------------------

// Subscriber listens on NATS subjects and delegates to the application service.
type Subscriber struct {
	nc          *nats.Conn
	userService *appuser.Service
	subs        []*nats.Subscription
}

// NewSubscriber creates a Subscriber bound to the given connection and service.
func NewSubscriber(nc *nats.Conn, userService *appuser.Service) *Subscriber {
	return &Subscriber{nc: nc, userService: userService}
}

// Subscribe registers all User handlers. Call Drain to shut down gracefully.
func (s *Subscriber) Subscribe() error {
	handlers := []struct {
		subject string
		handler nats.MsgHandler
	}{
		{SubjectCreateUser, s.handleCreate},
		{SubjectGetUser, s.handleGet},
		{SubjectListUsers, s.handleList},
		{SubjectUpdateUser, s.handleUpdate},
		{SubjectDeleteUser, s.handleDelete},
	}
	for _, h := range handlers {
		sub, err := s.nc.Subscribe(h.subject, h.handler)
		if err != nil {
			return err
		}
		s.subs = append(s.subs, sub)
		logger.L().Info("subscribed", zap.String("subject", h.subject))
	}
	return nil
}

// Drain unsubscribes and waits for all pending messages to be processed.
func (s *Subscriber) Drain() {
	for _, sub := range s.subs {
		_ = sub.Drain()
	}
}

// --- individual handlers ----------------------------------------------------

func (s *Subscriber) handleCreate(msg *nats.Msg) {
	var cmd appuser.CreateUserCommand
	if err := json.Unmarshal(msg.Data, &cmd); err != nil {
		replyErr(msg, "invalid request: "+err.Error())
		return
	}
	dto, err := s.userService.CreateUser(context.Background(), cmd)
	if err != nil {
		replyErr(msg, err.Error())
		return
	}
	replyOK(msg, dto)
}

func (s *Subscriber) handleGet(msg *nats.Msg) {
	var q appuser.GetUserQuery
	if err := json.Unmarshal(msg.Data, &q); err != nil {
		replyErr(msg, "invalid request: "+err.Error())
		return
	}
	dto, err := s.userService.GetUser(context.Background(), q)
	if err != nil {
		replyErr(msg, err.Error())
		return
	}
	replyOK(msg, dto)
}

func (s *Subscriber) handleList(msg *nats.Msg) {
	dtos, err := s.userService.ListUsers(context.Background(), appuser.ListUsersQuery{})
	if err != nil {
		replyErr(msg, err.Error())
		return
	}
	replyOK(msg, dtos)
}

func (s *Subscriber) handleUpdate(msg *nats.Msg) {
	var cmd appuser.UpdateUserCommand
	if err := json.Unmarshal(msg.Data, &cmd); err != nil {
		replyErr(msg, "invalid request: "+err.Error())
		return
	}
	dto, err := s.userService.UpdateUser(context.Background(), cmd)
	if err != nil {
		replyErr(msg, err.Error())
		return
	}
	replyOK(msg, dto)
}

func (s *Subscriber) handleDelete(msg *nats.Msg) {
	var cmd appuser.DeleteUserCommand
	if err := json.Unmarshal(msg.Data, &cmd); err != nil {
		replyErr(msg, "invalid request: "+err.Error())
		return
	}
	if err := s.userService.DeleteUser(context.Background(), cmd); err != nil {
		replyErr(msg, err.Error())
		return
	}
	replyOK(msg, map[string]string{"status": "ok"})
}
