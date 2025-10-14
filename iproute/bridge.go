package iproute

import (
	"github.com/pritunl/pritunl-cloud/utils"
)

func BridgeAdd(namespace, name string) (err error) {
	if namespace != "" {
		_, err = utils.ExecCombinedOutputLogged(
			[]string{
				"File exists",
			},
			"ip", "netns", "exec", namespace,
			"ip", "link",
			"add", name,
			"type", "bridge",
			"stp_state", "0",
			"forward_delay", "0",
		)
	} else {
		_, err = utils.ExecCombinedOutputLogged(
			[]string{
				"File exists",
			},
			"ip", "link",
			"add", name,
			"type", "bridge",
			"stp_state", "0",
			"forward_delay", "0",
		)
	}
	if err != nil {
		return
	}

	return
}

func BridgeDelete(namespace, name string) (err error) {
	if namespace != "" {
		_, err = utils.ExecCombinedOutputLogged(
			[]string{
				"File exists",
			},
			"ip", "netns", "exec", namespace,
			"ip", "link",
			"delete", name,
			"type", "bridge",
		)
	} else {
		_, err = utils.ExecCombinedOutputLogged(
			[]string{
				"File exists",
			},
			"ip", "link",
			"delete", name,
			"type", "bridge",
		)
	}
	if err != nil {
		return
	}

	return
}
