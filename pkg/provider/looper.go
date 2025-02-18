package provider

import (
	"github.com/pion/webrtc/v3"

	lksdk "github.com/dtelecom/server-sdk-go"
	"github.com/livekit/protocol/livekit"
)

type Looper interface {
	lksdk.SampleProvider
	Codec() webrtc.RTPCodecCapability
}

type VideoLooper interface {
	Looper
	ToLayer(quality livekit.VideoQuality) *livekit.VideoLayer
}
