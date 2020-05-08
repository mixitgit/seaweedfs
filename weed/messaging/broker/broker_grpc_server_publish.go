package broker

import (
	"io"

	"github.com/golang/protobuf/proto"

	"github.com/chrislusf/seaweedfs/weed/glog"
	"github.com/chrislusf/seaweedfs/weed/pb/messaging_pb"
)

func (broker *MessageBroker) Publish(stream messaging_pb.SeaweedMessaging_PublishServer) error {

	// process initial request
	in, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}

	// TODO look it up
	topicConfig := &messaging_pb.TopicConfiguration{
		// IsTransient: true,
	}

	// send init response
	initResponse := &messaging_pb.PublishResponse{
		Config:   nil,
		Redirect: nil,
	}
	err = stream.Send(initResponse)
	if err != nil {
		return err
	}
	if initResponse.Redirect != nil {
		return nil
	}

	// get lock
	tp := TopicPartition{
		Namespace: in.Init.Namespace,
		Topic:     in.Init.Topic,
		Partition: in.Init.Partition,
	}
	tl := broker.topicLocks.RequestLock(tp, topicConfig, true)
	defer broker.topicLocks.ReleaseLock(tp, true)

	// process each message
	for {
		// println("recv")
		in, err := stream.Recv()
		// glog.V(0).Infof("recieved %v err: %v", in, err)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		if in.Data == nil {
			continue
		}

		// fmt.Printf("received: %d : %s\n", len(in.Data.Value), string(in.Data.Value))

		data, err := proto.Marshal(in.Data)
		if err != nil {
			glog.Errorf("marshall error: %v\n", err)
			continue
		}

		tl.logBuffer.AddToBuffer(in.Data.Key, data)

		if in.Data.IsClose {
			// println("server received closing")
			break
		}

	}

	// send the close ack
	// println("server send ack closing")
	if err := stream.Send(&messaging_pb.PublishResponse{IsClosed: true}); err != nil {
		glog.V(0).Infof("err sending close response: %v", err)
	}
	return nil

}
