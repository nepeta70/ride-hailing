package grpc

import (
	"context"

	notificationv1 "github.com/nepeta70/ride-hailing/gen/proto/notification/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

// NotificationHandler implements the NotificationService gRPC interface.
type NotificationHandler struct {
	notificationv1.UnimplementedNotificationServiceServer
	// Add service dependency here, e.g. notificationService *service.NotificationService
}

func NewNotificationHandler( /* service *service.NotificationService */ ) *NotificationHandler {
	return &NotificationHandler{ /* notificationService: service */ }
}

func (h *NotificationHandler) SendNotification(ctx context.Context, req *notificationv1.SendNotificationRequest) (*emptypb.Empty, error) {
	// TODO: Call the notification service logic here
	// Example:
	// err := h.notificationService.Send(ctx, req.UserId, req.Title, req.Message, req.Data)
	// if err != nil {
	//     return nil, err
	// }
	return &emptypb.Empty{}, nil
}
