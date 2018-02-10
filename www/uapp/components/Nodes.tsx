/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as NodeTypes from '../types/NodeTypes';
import * as CertificateTypes from '../types/CertificateTypes';
import NodesStore from '../stores/NodesStore';
import CertificatesStore from '../stores/CertificatesStore';
import * as NodeActions from '../actions/NodeActions';
import * as CertificateActions from '../actions/CertificateActions';
import Node from './Node';
import Page from './Page';
import PageHeader from './PageHeader';

interface State {
	nodes: NodeTypes.NodesRo;
	certificates: CertificateTypes.CertificatesRo;
	disabled: boolean;
}

const css = {
	header: {
		marginTop: '-19px',
	} as React.CSSProperties,
	heading: {
		margin: '19px 0 0 0',
	} as React.CSSProperties,
};

export default class Nodes extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			nodes: NodesStore.nodes,
			certificates: CertificatesStore.certificates,
			disabled: false,
		};
	}

	componentDidMount(): void {
		NodesStore.addChangeListener(this.onChange);
		CertificatesStore.addChangeListener(this.onChange);
		NodeActions.sync();
		CertificateActions.sync();
	}

	componentWillUnmount(): void {
		NodesStore.removeChangeListener(this.onChange);
		CertificatesStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		this.setState({
			...this.state,
			nodes: NodesStore.nodes,
			certificates: CertificatesStore.certificates,
		});
	}

	render(): JSX.Element {
		let nodesDom: JSX.Element[] = [];

		this.state.nodes.forEach((node: NodeTypes.NodeRo): void => {
			nodesDom.push(<Node
				key={node.id}
				node={node}
				certificates={this.state.certificates}
			/>);
		});

		return <Page>
			<PageHeader>
				<div className="layout horizontal wrap" style={css.header}>
					<h2 style={css.heading}>Nodes</h2>
					<div className="flex"/>
				</div>
			</PageHeader>
			<div>
				{nodesDom}
			</div>
		</Page>;
	}
}
