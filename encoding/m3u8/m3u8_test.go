package m3u8

import "testing"

func TestParse_SimpleMediaPlaylist(t *testing.T) {
	body := []byte(`#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:10
#EXT-X-MEDIA-SEQUENCE:0
#EXTINF:10.0,
http://example.com/segment1.ts
#EXTINF:10.0,
http://example.com/segment2.ts
#EndList
`)

	m, err := Parse(body)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if m.Version != 3 {
		t.Errorf("Version = %d, want 3", m.Version)
	}
	if m.TargetDuration != 10 {
		t.Errorf("TargetDuration = %f, want 10", m.TargetDuration)
	}
	if m.MediaSequence != 0 {
		t.Errorf("MediaSequence = %d, want 0", m.MediaSequence)
	}
	if len(m.Segments) != 2 {
		t.Fatalf("Segments count = %d, want 2", len(m.Segments))
	}
	if m.Segments[0].URI != "http://example.com/segment1.ts" {
		t.Errorf("Segment[0].URI = %q, want segment1.ts", m.Segments[0].URI)
	}
	if m.Segments[0].Duration != 10.0 {
		t.Errorf("Segment[0].Duration = %f, want 10.0", m.Segments[0].Duration)
	}
	if !m.EndList {
		t.Error("EndList should be true")
	}
}

func TestParse_MasterPlaylist(t *testing.T) {
	body := []byte(`#EXTM3U
#EXT-X-STREAM-INF:BANDWIDTH=240000,RESOLUTION=416x234,CODECS="avc1.42e00a,mp4a.40.2"
http://example.com/low/index.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=800000,RESOLUTION=864x480
http://example.com/high/index.m3u8
`)

	m, err := Parse(body)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(m.MasterPlaylist) != 2 {
		t.Fatalf("MasterPlaylist count = %d, want 2", len(m.MasterPlaylist))
	}
	if m.MasterPlaylist[0].BandWidth != 240000 {
		t.Errorf("BandWidth = %d, want 240000", m.MasterPlaylist[0].BandWidth)
	}
	if m.MasterPlaylist[0].Resolution != "416x234" {
		t.Errorf("Resolution = %q, want 416x234", m.MasterPlaylist[0].Resolution)
	}
}

func TestParse_InvalidHeader(t *testing.T) {
	body := []byte("not an m3u8 file")
	_, err := Parse(body)
	if err == nil {
		t.Error("Parse() should return error for invalid m3u8")
	}
}

func TestParse_WithKey(t *testing.T) {
	body := []byte(`#EXTM3U
#EXT-X-KEY:METHOD=AES-128,URI="key.key"
#EXTINF:10.0,
http://example.com/segment1.ts
`)

	m, err := Parse(body)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(m.Keys) == 0 {
		t.Fatal("Keys should not be empty")
	}
	for _, k := range m.Keys {
		if k.Method != CryptMethodAES {
			t.Errorf("Key Method = %q, want AES-128", k.Method)
		}
		if k.URI != "key.key" {
			t.Errorf("Key URI = %q, want key.key", k.URI)
		}
	}
}

func TestParse_PlaylistType(t *testing.T) {
	body := []byte(`#EXTM3U
#EXT-X-PLAYLIST-TYPE:VOD
#EXTINF:10.0,
http://example.com/segment1.ts
`)

	m, err := Parse(body)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if m.PlaylistType != PlaylistTypeVOD {
		t.Errorf("PlaylistType = %q, want VOD", m.PlaylistType)
	}
}

func TestParse_ByteRange(t *testing.T) {
	body := []byte(`#EXTM3U
#EXTINF:10.0,
#EXT-X-BYTERANGE:1000@500
http://example.com/segment1.ts
`)

	m, err := Parse(body)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(m.Segments) != 1 {
		t.Fatalf("Segments count = %d, want 1", len(m.Segments))
	}
	if m.Segments[0].Length != 1000 {
		t.Errorf("Segment Length = %d, want 1000", m.Segments[0].Length)
	}
	if m.Segments[0].Offset != 500 {
		t.Errorf("Segment Offset = %d, want 500", m.Segments[0].Offset)
	}
}
