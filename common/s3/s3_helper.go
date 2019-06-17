package s3

import (
	"errors"
	"fmt"
	"net/http"
	"path"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/sirupsen/logrus"
	"github.com/cenkalti/backoff"
)

// CopyImageToS3FromURL takes an image at some original location and re-uploads it to S3 in the given bucket
func CopyImageToS3FromURL(productLogger *logrus.Entry, uploader s3manageriface.UploaderAPI, sourceURL string, destBucket string) (string, error) {
	if sourceURL == "" {
		return "", errors.New("url cannot be empty")
	}

	if uploader == nil {
		return "", errors.New("uploader cannot be nil")
	}

	var resp *http.Response
	err := backoff.Retry(func() error {
		resp, err := http.Get(sourceURL)
		if err != nil || (resp != nil && resp.StatusCode > 299) {
			productLogger.WithField("originalImageURL", sourceURL).WithError(err).Warning("Could not download the product image from the image URL specified, retrying...")
			return err
		}
		return nil
	}, backoff.NewExponentialBackOff())

	if err != nil {
		productLogger.WithField("originalImageURL", sourceURL).WithError(err).Error("Could not download the product image from the image URL specified after exhausting the exponential backoff policy")
		return "", err
	}

	s3Key := fmt.Sprintf("img/products/%s", path.Base(sourceURL))
	s3Logger := productLogger.WithField("s3Key", s3Key)
	s3Logger.Info("Uploading product image to S3")
	s3Resp, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: &destBucket,
		Key:    &s3Key,
		Body:   resp.Body,
	})
	if err != nil {
		return "", err
	}

	s3Logger.WithField("s3UploadLocation", s3Resp.Location).Info("Finished uploading product image to S3")
	resp.Body.Close()

	return s3Key, nil
}
