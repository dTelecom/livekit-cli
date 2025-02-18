package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"

	lksdk "github.com/dtelecom/server-sdk-go"
	"github.com/livekit/protocol/livekit"
)

const roomCategory = "Room Server API"

var (
	RoomCommands = []*cli.Command{
		{
			Name:     "delete-room",
			Before:   createRoomClient,
			Action:   deleteRoom,
			Category: roomCategory,
			Flags: withDefaultFlags(
				roomFlag,
			),
		},
		{
			Name:     "remove-participant",
			Before:   createRoomClient,
			Action:   removeParticipant,
			Category: roomCategory,
			Flags: withDefaultFlags(
				roomFlag,
				identityFlag,
			),
		},
		{
			Name:     "mute-track",
			Before:   createRoomClient,
			Action:   muteTrack,
			Category: roomCategory,
			Flags: withDefaultFlags(
				roomFlag,
				identityFlag,
				&cli.StringFlag{
					Name:     "track",
					Usage:    "track sid to mute",
					Required: true,
				},
				&cli.BoolFlag{
					Name:  "muted",
					Usage: "set to true to mute, false to unmute",
				},
			),
		},
	}

	roomClient *lksdk.RoomServiceClient
)

func createRoomClient(c *cli.Context) error {
	pc, err := loadProjectDetails(c)
	if err != nil {
		return err
	}

	roomClient = lksdk.NewRoomServiceClient(pc.URL, pc.APIKey, pc.APISecret)
	return nil
}

func deleteRoom(c *cli.Context) error {
	roomId := c.String("room")
	_, err := roomClient.DeleteRoom(context.Background(), &livekit.DeleteRoomRequest{
		Room: roomId,
	})
	if err != nil {
		return err
	}

	fmt.Println("deleted room", roomId)
	return nil
}

func removeParticipant(c *cli.Context) error {
	roomName, identity := participantInfoFromCli(c)
	_, err := roomClient.RemoveParticipant(context.Background(), &livekit.RoomParticipantIdentity{
		Room:     roomName,
		Identity: identity,
	})
	if err != nil {
		return err
	}

	fmt.Println("successfully removed participant", identity)

	return nil
}

func muteTrack(c *cli.Context) error {
	roomName, identity := participantInfoFromCli(c)
	trackSid := c.String("track")
	_, err := roomClient.MutePublishedTrack(context.Background(), &livekit.MuteRoomTrackRequest{
		Room:     roomName,
		Identity: identity,
		TrackSid: trackSid,
		Muted:    c.Bool("muted"),
	})
	if err != nil {
		return err
	}

	verb := "muted"
	if !c.Bool("muted") {
		verb = "unmuted"
	}
	fmt.Println(verb, "track: ", trackSid)
	return nil
}

func participantInfoFromCli(c *cli.Context) (string, string) {
	return c.String("room"), c.String("identity")
}
