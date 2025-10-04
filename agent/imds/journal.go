package imds

import (
	"github.com/pritunl/pritunl-cloud/agent/logging"
	"github.com/pritunl/tools/logger"
)

type Journal struct {
	Index   int32           `json:"-"`
	Key     string          `json:"-"`
	Type    string          `json:"-"`
	Unit    string          `json:"-"`
	Path    string          `json:"-"`
	Handler logging.Handler `json:"-"`
}

func (j *Journal) Open() (err error) {
	logger.WithFields(logger.Fields{
		"index": j.Index,
		"key":   j.Key,
		"type":  j.Type,
		"unit":  j.Unit,
		"path":  j.Path,
	}).Info("agent: Starting journal")

	err = j.Handler.Open()
	if err != nil {
		return
	}

	return
}

func (j *Journal) Close() {
	logger.WithFields(logger.Fields{
		"index": j.Index,
		"key":   j.Key,
		"type":  j.Type,
		"unit":  j.Unit,
		"path":  j.Path,
	}).Info("agent: Stopping journal")

	err := j.Handler.Close()
	if err != nil {
		logger.WithFields(logger.Fields{
			"index": j.Index,
			"key":   j.Key,
			"type":  j.Type,
			"unit":  j.Unit,
			"path":  j.Path,
			"error": err,
		}).Error("agent: Error stopping journal")
	}

	return
}
