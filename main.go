package main

import (
	"context"
	"fmt"
	"os"
	"path"

	//"go/types"
	"log"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var output4K bool

type OutType string

const (
	OUT_4K   OutType = "4K"
	OUT_1080 OutType = "HD"
)

type Resolution struct {
	Width  int
	Height int
}

type Output struct {
	outType OutType
	Res     Resolution
}

func main() {

	errs := MakeVideos([]OutType{})
	for _, err := range errs {
		fmt.Printf("error: %s\n", err.Error())
	}
}

func MakeVideos(outTypes []OutType) []error {

	errs := []error{}
	for _, ot := range outTypes {
		err := makeVideo(ot)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func makeVideo(outType OutType) error {
	rootDir := "/videos" //"/Users/tom/Videos/theater_demos" //
	files, err := os.ReadDir(rootDir)
	if err != nil {
		return err
	}

	outputType := Output{
		outType: outType,
	}
	switch outType {
	case OUT_1080:
		outputType.Res = Resolution{
			Width:  1920,
			Height: 1080,
		}
	case OUT_4K:
		outputType.Res = Resolution{
			Width:  3840,
			Height: 2160,
		}
	default:
		return fmt.Errorf("unhandled OutType %s", outType)
	}

	complex := ""
	complexOut := ""
	cmd := []string{"-y"}
	videos := 0
	for _, file := range files {
		if file.Name() == ".DS_Store" {
			continue
		}
		cmd = append(cmd, "-i")
		cmd = append(cmd, path.Join(rootDir, file.Name()))
		complex += fmt.Sprintf("[%d:v]scale=%d:%d,pad=1280:ih:(ow-iw)/2[v%d]; ", videos, outputType.Res.Width, outputType.Res.Height, videos)
		complex += fmt.Sprintf("[%d:a]aformat=sample_fmts=s32:sample_rates=48000[a%d]; ", videos, videos)
		// aformat=sample_fmts=s32:sample_rates=48000[a];[a]channelsplit=channel_layout=stereo[FL][FR]
		complexOut += fmt.Sprintf("[v%d][a%d]", videos, videos)
		videos++
	}
	cmd = append(cmd, "-c:v", "libx264", "-pix_fmt", "yuv420p", "-r", "60", "-c:a", "aac", "-ac", "2", "-ar", "48000")
	//cmd = append(cmd, "-vf", "scale=1920:1080")
	complex += complexOut
	complex += fmt.Sprintf("concat=n=%d:v=1:a=1 [vv] [aa]", videos)
	cmd = append(cmd, "-filter_complex", complex)
	cmd = append(cmd, "-map", "[vv]", "-map", "[aa]")
	cmd = append(cmd, "-movflags", "+faststart") // Put the  MOOV atom at the beginning so FFpeobe can quickly parse it.
	cmd = append(cmd, "theaterdemos.mp4")
	fmt.Printf("%v\n", cmd)
	out, err := exec.Command("ffmpeg", cmd...).CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Printf("%s", out)
	return nil
}

// BucketBasics encapsulates the Amazon Simple Storage Service (Amazon S3) actions
// used in the examples.
// It contains S3Client, an Amazon S3 service client that is used to perform bucket
// and object actions.
type BucketBasics struct {
	S3Client *s3.Client
}

// ListObjects lists the objects in a bucket.
func (basics BucketBasics) ListObjects(bucketName string) ([]types.Object, error) {
	result, err := basics.S3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	var contents []types.Object
	if err != nil {
		log.Printf("Couldn't list objects in bucket %v. Here's why: %v\n", bucketName, err)
	} else {
		contents = result.Contents
	}
	return contents, err
}
