package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	//"go/types"
	"log"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type OutType string

const (
	OUT_4K   OutType = "4K"
	OUT_1080 OutType = "HD"
)

var outFileName string = "theaterdemos"
var fast *bool
var short *bool

type Resolution struct {
	Width  int
	Height int
}

type Output struct {
	outType OutType
	Res     Resolution
}

func main() {

	res := flag.String("res", "HD", "Comma-delimited list of output resolutions. 'HD' and '4K'.")
	useS3 := flag.Bool("s3", false, "If true, input videos will come from S3. If false, local ~/Videos folder.")
	fast = flag.Bool("fast", false, "Low quality, ultrafast - for testing config")
	short = flag.Bool("short", false, "Only encode ten seconds")

	flag.Parse()

	fmt.Printf("res %s useS3 %t fast %t\n", *res, *useS3, *fast)

	reses := strings.Split(*res, ",")
	outTypes := []OutType{}
	for _, ot := range reses {
		outTypes = append(outTypes, OutType(ot))
	}

	errs := MakeVideos(outTypes, *useS3)
	for _, err := range errs {
		fmt.Printf("error: %s\n", err.Error())
	}
}

func MakeVideos(outTypes []OutType, useS3 bool) []error {

	errs := []error{}
	var err error
	var paths []string

	if useS3 {
		p := getS3Videos()
		for _, pa := range p {
			if _, err := os.Stat(path.Base(pa)); err != nil {
				r, err := exec.Command("wget", pa).CombinedOutput()
				if err != nil {
					return append(errs, err)
				}
				fmt.Printf("%s", r)
			}
			paths = append(paths, path.Base(pa))

		}
	} else {
		paths, err = getLocalVideoList()
	}
	if err != nil {
		errs = append(errs, err)
		return errs
	}

	for _, ot := range outTypes {
		err := makeVideo(ot, paths)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func getLocalVideoList() ([]string, error) {
	rootDir := "/videos" //"/Users/tom/Videos/theater_demos" //
	paths := []string{}
	files, err := os.ReadDir(rootDir)
	if err != nil {
		return paths, err
	}
	for _, file := range files {
		paths = append(paths, path.Join(rootDir, file.Name()))
	}
	return paths, nil
}

func makeVideo(outType OutType, files []string) error {

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
		if strings.Contains(file, ".DS_Store") {
			continue
		}
		cmd = append(cmd, "-i")
		cmd = append(cmd, file)
		complex += fmt.Sprintf(`[%d:v]colorspace=bt709:iall=bt601-6-625:fast=1,scale=%d:%d:force_original_aspect_ratio=1,pad=%d:%d:(ow-iw)/2:(oh-ih)/2,drawtext=fontfile='./assets/fonts/AzeretMono.ttf':text='':
		x=(w-tw-60):y=(h-th-60):fontsize=40:fontcolor=white:timecode='00\:00\:00\:00':timecode_rate=%d[v%d]; `, videos, outputType.Res.Width, outputType.Res.Height, outputType.Res.Width, outputType.Res.Height, 60, videos)
		complex += fmt.Sprintf("[%d:a]aformat=sample_fmts=s32:sample_rates=48000[a%d]; ", videos, videos)
		// aformat=sample_fmts=s32:sample_rates=48000[a];[a]channelsplit=channel_layout=stereo[FL][FR]
		complexOut += fmt.Sprintf("[v%d][a%d]", videos, videos)
		videos++
	}
	cmd = append(cmd, "-c:v", "libx264", "-pix_fmt", "yuv420p", "-r", "60", "-c:a", "libfdk_aac", "-b:a", "256k", "-ac", "2", "-ar", "48000")
	cmd = append(cmd, "-sws_flags", "spline+accurate_rnd+full_chroma_int")
	cmd = append(cmd, "-color_range", "1", "-colorspace", "1", "-color_primaries", "1", "-color_trc", "1")
	if *fast {
		outFileName += "-fast"
		cmd = append(cmd, "-preset", "ultrafast")
	} else {
		cmd = append(cmd, "-preset", "veryslow", "-crf", "17")
	}
	if *short {
		outFileName += "-short"
		cmd = append(cmd, "-t", "10")
	}

	//cmd = append(cmd, "-vf", "scale=1920:1080")
	complex += complexOut
	complex += fmt.Sprintf("concat=n=%d:v=1:a=1 [vv] [aa]", videos)
	cmd = append(cmd, "-filter_complex", complex)
	cmd = append(cmd, "-map", "[vv]", "-map", "[aa]")
	cmd = append(cmd, "-movflags", "+faststart") // Put the  MOOV atom at the beginning so FFpeobe can quickly parse it.
	cmd = append(cmd, outFileName+string(outputType.outType)+".mp4")
	//fmt.Printf("%v\n", cmd)
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

func getS3Videos() []string {
	files := []string{
		// "https://batch-ffmpeg-stack-batchffmpegbucketd97ee012-mkr8qp9ts9jd.s3.us-west-2.amazonaws.com/4K-theater-demos/LG 4K HDR Demo - New York.ts",

		// "https://batch-ffmpeg-stack-batchffmpegbucketd97ee012-mkr8qp9ts9jd.s3.us-west-2.amazonaws.com/4K-theater-demos/Samsung-Ride-on-Board-4K-(www.demolandia.net).ts",

		// "https://batch-ffmpeg-stack-batchffmpegbucketd97ee012-mkr8qp9ts9jd.s3.us-west-2.amazonaws.com/4K-theater-demos/Samsung-and-RedBull-See-the-Unexpected-HDR-UHD-4K-(www.demolandia.net).ts",

		//"https://batch-ffmpeg-stack-batchffmpegbucketd97ee012-mkr8qp9ts9jd.s3.us-west-2.amazonaws.com/4K-theater-demos/gopro1.mp4",

		"https://batch-ffmpeg-stack-batchffmpegbucketd97ee012-mkr8qp9ts9jd.s3.us-west-2.amazonaws.com/4K-theater-demos/Sony-Food-Fizzle-UHD-4K-(www.demolandia.net).mp4",

		"https://batch-ffmpeg-stack-batchffmpegbucketd97ee012-mkr8qp9ts9jd.s3.us-west-2.amazonaws.com/4K-theater-demos/Sony 4K HDR Demo - New York Fashion.mp4",

		"https://batch-ffmpeg-stack-batchffmpegbucketd97ee012-mkr8qp9ts9jd.s3.us-west-2.amazonaws.com/4K-theater-demos/imax-cliffhanger-flat-(www.demolandia.net).mkv",

		"https://batch-ffmpeg-stack-batchffmpegbucketd97ee012-mkr8qp9ts9jd.s3.us-west-2.amazonaws.com/4K-theater-demos/dolby-chameleon-uhd-(www.demolandia.net).mkv",

		"https://batch-ffmpeg-stack-batchffmpegbucketd97ee012-mkr8qp9ts9jd.s3.us-west-2.amazonaws.com/4K-theater-demos/Sony-Swordsmith-HDR-UHD-4K-(www.demolandia.net).mp4",

		"https://batch-ffmpeg-stack-batchffmpegbucketd97ee012-mkr8qp9ts9jd.s3.us-west-2.amazonaws.com/4K-theater-demos/bbb_sunflower_2160p_60fps_normal.mp4",

		"https://batch-ffmpeg-stack-batchffmpegbucketd97ee012-mkr8qp9ts9jd.s3.us-west-2.amazonaws.com/4K-theater-demos/tearsofsteel_4k.mov",
	}
	return files
}
