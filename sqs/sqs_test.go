package sqs_test

import (
	"os"
	"testing"

	"github.com/kickback-app/common/sqs"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestInit(t *testing.T) {
	_, err := sqs.NewClient(&sqs.ClientParams{
		Region:       "us-east-1",
		AccessKey:    "",
		AccessSecret: "aaa",
		QueueName:    "test",
	})
	require.NotNil(t, err, "there should be an err if no access key")
}
