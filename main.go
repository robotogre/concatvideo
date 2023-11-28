package main

import (
	"context"
	"fmt"

	//"go/types"
	"log"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func main() {

	// cfg, err := config.LoadDefaultConfig(context.TODO(),
	// 	config.WithRegion("us-west-2"),
	// )
	// if err != nil {
	// 	// handle error
	// }
	// client := s3.NewFromConfig(cfg)
	// bb := BucketBasics{
	// 	S3Client: client,
	// }
	// obs, err := bb.ListObjects("batch-ffmpeg-stack-batchffmpegbucketd97ee012-mkr8qp9ts9jd")

	// for _, obj := range obs {
	// 	fmt.Printf("%v\n", obj.Key)
	// }
	// return
	// rootDir := "/Users/tom/Downloads/theater_demos"
	// files, err := os.ReadDir(rootDir)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	files := []string{
		// "https://batch-ffmpeg-stack-batchffmpegbucketd97ee012-mkr8qp9ts9jd.s3.us-west-2.amazonaws.com/4K-theater-demos/LG 4K HDR Demo - New York.ts",

		// "https://batch-ffmpeg-stack-batchffmpegbucketd97ee012-mkr8qp9ts9jd.s3.us-west-2.amazonaws.com/4K-theater-demos/Samsung-Ride-on-Board-4K-(www.demolandia.net).ts",

		// "https://batch-ffmpeg-stack-batchffmpegbucketd97ee012-mkr8qp9ts9jd.s3.us-west-2.amazonaws.com/4K-theater-demos/Samsung-and-RedBull-See-the-Unexpected-HDR-UHD-4K-(www.demolandia.net).ts",

		"https://batch-ffmpeg-stack-batchffmpegbucketd97ee012-mkr8qp9ts9jd.s3.us-west-2.amazonaws.com/4K-theater-demos/Sony 4K HDR Demo - New York Fashion.mp4",

		"https://batch-ffmpeg-stack-batchffmpegbucketd97ee012-mkr8qp9ts9jd.s3.us-west-2.amazonaws.com/4K-theater-demos/Sony-Food-Fizzle-UHD-4K-(www.demolandia.net).mp4",

		"https://batch-ffmpeg-stack-batchffmpegbucketd97ee012-mkr8qp9ts9jd.s3.us-west-2.amazonaws.com/4K-theater-demos/Sony-Swordsmith-HDR-UHD-4K-(www.demolandia.net).mp4",

		"https://batch-ffmpeg-stack-batchffmpegbucketd97ee012-mkr8qp9ts9jd.s3.us-west-2.amazonaws.com/4K-theater-demos/bbb_sunflower_2160p_60fps_normal.mp4",

		"https://batch-ffmpeg-stack-batchffmpegbucketd97ee012-mkr8qp9ts9jd.s3.us-west-2.amazonaws.com/4K-theater-demos/dolby-chameleon-uhd-(www.demolandia.net).mkv",

		"https://batch-ffmpeg-stack-batchffmpegbucketd97ee012-mkr8qp9ts9jd.s3.us-west-2.amazonaws.com/4K-theater-demos/gopro1.mp4",

		"https://batch-ffmpeg-stack-batchffmpegbucketd97ee012-mkr8qp9ts9jd.s3.us-west-2.amazonaws.com/4K-theater-demos/imax-cliffhanger-flat-(www.demolandia.net).mkv",

		"https://batch-ffmpeg-stack-batchffmpegbucketd97ee012-mkr8qp9ts9jd.s3.us-west-2.amazonaws.com/4K-theater-demos/tearsofsteel_4k.mov",
	}

	complex := ""
	complexOut := ""
	cmd := []string{"-y"}
	videos := 0
	for _, file := range files {
		// if file.Name() == ".DS_Store" {
		// 	continue
		// }
		cmd = append(cmd, "-i")
		cmd = append(cmd, file)
		complex += fmt.Sprintf("[%d:v]scale=1920:1080,setdar=16/9[v%d]; ", videos, videos)
		complexOut += fmt.Sprintf("[v%d][%d:a]", videos, videos)
		videos++
	}
	cmd = append(cmd, "-c:v", "libx264", "-pix_fmt", "yuv420p", "-r", "60", "-c:a", "aac", "-ac", "2")
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
		//log.Fatal(err)
	}
	fmt.Printf("%s", out)

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
