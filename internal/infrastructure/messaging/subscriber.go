package messaging

import (
	"context"

	appuser "github.com/chengtb/chengtb/internal/application/user"
	"github.com/chengtb/chengtb/internal/pb"
	"github.com/chengtb/chengtb/pkg/logger"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
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

// replyOK marshals payload as a protobuf message and sends a success Reply.
func replyOK(msg *nats.Msg, payload proto.Message) {
	var data []byte
	if payload != nil {
		var err error
		data, err = proto.Marshal(payload)
		if err != nil {
			replyErr(msg, 1, "internal error: marshal failed")
			return
		}
	}
	envelope, err := proto.Marshal(&pb.Reply{ErrorCode: 0, Data: data})
	if err != nil {
		logger.L().Error("nats marshal reply envelope", zap.Error(err))
		return
	}
	if err := msg.Respond(envelope); err != nil {
		logger.L().Error("nats respond error", zap.Error(err))
	}
}

// replyErr sends a Reply with the given error code and message.
func replyErr(msg *nats.Msg, code int32, errMsg string) {
	envelope, err := proto.Marshal(&pb.Reply{
		ErrorCode:    code,
		ErrorMessage: errMsg,
	})
	if err != nil {
		logger.L().Error("nats marshal error envelope", zap.Error(err))
		return
	}
	if err := msg.Respond(envelope); err != nil {
		logger.L().Error("nats respond error", zap.Error(err))
	}
}

// userToProto converts an appuser.UserDTO to a pb.User.
func userToProto(dto *appuser.UserDTO) *pb.User {
	return &pb.User{
		Id:            dto.ID,
		Name:          dto.Name,
		Email:         dto.Email,
		CreatedAtUnix: dto.CreatedAt.UnixNano(),
		UpdatedAtUnix: dto.UpdatedAt.UnixNano(),
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
	var req pb.CreateUserRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		replyErr(msg, 1, "invalid request: "+err.Error())
		return
	}
	dto, err := s.userService.CreateUser(context.Background(), appuser.CreateUserCommand{
		ID:    req.Id,
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		replyErr(msg, 1, err.Error())
		return
	}
	replyOK(msg, userToProto(dto))
}

func (s *Subscriber) handleGet(msg *nats.Msg) {
	var req pb.GetUserRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		replyErr(msg, 1, "invalid request: "+err.Error())
		return
	}
	dto, err := s.userService.GetUser(context.Background(), appuser.GetUserQuery{ID: req.Id})
	if err != nil {
		replyErr(msg, 1, err.Error())
		return
	}
	replyOK(msg, userToProto(dto))
}

func (s *Subscriber) handleList(msg *nats.Msg) {
	var req pb.ListUsersRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		replyErr(msg, 1, "invalid request: "+err.Error())
		return
	}
	dtos, err := s.userService.ListUsers(context.Background(), appuser.ListUsersQuery{})
	if err != nil {
		replyErr(msg, 1, err.Error())
		return
	}
	users := make([]*pb.User, 0, len(dtos))
	for _, d := range dtos {
		users = append(users, userToProto(d))
	}
	replyOK(msg, &pb.UserList{Users: users})
}

func (s *Subscriber) handleUpdate(msg *nats.Msg) {
	var req pb.UpdateUserRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		replyErr(msg, 1, "invalid request: "+err.Error())
		return
	}
	dto, err := s.userService.UpdateUser(context.Background(), appuser.UpdateUserCommand{
		ID:    req.Id,
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		replyErr(msg, 1, err.Error())
		return
	}
	replyOK(msg, userToProto(dto))
}

func (s *Subscriber) handleDelete(msg *nats.Msg) {
	var req pb.DeleteUserRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		replyErr(msg, 1, "invalid request: "+err.Error())
		return
	}
	if err := s.userService.DeleteUser(context.Background(), appuser.DeleteUserCommand{ID: req.Id}); err != nil {
		replyErr(msg, 1, err.Error())
		return
	}
	replyOK(msg, &pb.Empty{})
}
