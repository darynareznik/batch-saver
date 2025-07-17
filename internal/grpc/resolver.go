package grpc

import (
	"batch-saver/api"
	"batch-saver/internal/models"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
)

type service interface {
	Save(e models.Event)
}

type resolver struct {
	service service
	api.UnimplementedBatchSaverServiceServer
}

func NewResolver(service service) *resolver {
	return &resolver{service: service}
}

func (r *resolver) SaveEvents(stream api.BatchSaverService_SaveEventsServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&api.SaveEventsResponse{})
		}
		if err != nil {
			log.Error().Err(err).Msg("Error reading from a stream")
			return err
		}

		log.Debug().
			Any("event", req.Event).
			Msg("Got event")

		e, err := models.EventFromAPI(req.Event)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		r.service.Save(e)
	}
}
