package main

import "github.com/reconquest/hierr-go"

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

	tracef(`starting sync protocol with %d nodes`, len(execution.nodes))

	err = protocol.Init(execution.stdin)
	if err != nil {
		return hierr.Errorf(
			err,
			`can't init protocol with sync tool`,
		)
	}

	tracef(`sending information about %d nodes to each`, len(execution.nodes))

	nodes := []*remoteExecutionNode{}
	for _, node := range execution.nodes {
		nodes = append(nodes, node)
	}

	for _, node := range execution.nodes {
		for _, neighbor := range nodes {
			err := protocol.SendNode(node, neighbor)
			if err != nil {
				return hierr.Errorf(
					err,
					`can't send node to sync tool: '%s'`,
					node.String(),
				)
			}
		}
	}

	tracef(`sending start message to sync tools`)

	err = protocol.SendStart()
	if err != nil {
		return hierr.Errorf(
			err,
			`can't start sync tool`,
		)
	}

	debugf(`waiting sync tool to finish`)

	err = execution.wait()
	if err != nil {
		return hierr.Errorf(
			err,
			`failed to finish sync tool command`,
		)
	}

	return nil
}
