package main

import "github.com/seletskiy/hierr"

func runSyncProtocol(
	cluster *distributedLock,
	runner *remoteExecutionRunner,
) error {
	protocol := newSyncProtocol()

	execution, err := runner.run(
		cluster,
		func(remoteNode *remoteExecutionNode) {
			remoteNode.stdout = newProtocolNodeWriter(remoteNode, protocol)
		},
	)
	if err != nil {
		return hierr.Errorf(
			err,
			`can't run sync tool command`,
		)
	}

	err = protocol.Init(execution.stdin)
	if err != nil {
		return hierr.Errorf(
			err,
			`can't init protocol with sync tool`,
		)
	}

	for _, node := range execution.nodes {
		err := protocol.SendNode(node)
		if err != nil {
			return hierr.Errorf(
				err,
				`can't send node to sync tool: '%s'`,
				node.String(),
			)
		}
	}

	err = protocol.SendStart()
	if err != nil {
		return hierr.Errorf(
			err,
			`can't start sync tool`,
		)
	}

	err = execution.wait()
	if err != nil {
		return hierr.Errorf(
			err,
			`failed to finish sync tool command`,
		)
	}

	return nil
}
