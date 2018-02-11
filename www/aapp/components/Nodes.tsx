/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as NodeTypes from '../types/NodeTypes';
import * as CertificateTypes from '../types/CertificateTypes';
import NodesStore from '../stores/NodesStore';
import CertificatesStore from '../stores/CertificatesStore';
import * as NodeActions from '../actions/NodeActions';
import * as CertificateActions from '../actions/CertificateActions';
import Node from './Node';
import NodesPage from './NodesPage';
import Page from './Page';
import PageHeader from './PageHeader';

interface Selected {
	[key: string]: boolean;
}

interface State {
	nodes: NodeTypes.NodesRo;
	certificates: CertificateTypes.CertificatesRo;
	selected: Selected;
	lastSelected: string;
	disabled: boolean;
}

const css = {
	items: {
		width: '100%',
		marginTop: '-5px',
		display: 'table',
		borderSpacing: '0 5px',
	} as React.CSSProperties,
	itemsBox: {
		width: '100%',
		overflowY: 'auto',
	} as React.CSSProperties,
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
			selected: {},
			lastSelected: null,
			disabled: false,
		};
	}

	get selected(): boolean {
		for (let key in this.state.selected) {
			if (this.state.selected[key]) {
				return true;
			}
		}
		return false;
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
		let selected: Selected = {};
		let curSelected = this.state.selected;

		this.state.nodes.forEach((node: NodeTypes.Node): void => {
			if (curSelected[node.id]) {
				selected[node.id] = true;
			}
		});

		this.setState({
			...this.state,
			nodes: NodesStore.nodes,
			certificates: CertificatesStore.certificates,
			selected: selected,
		});
	}

	render(): JSX.Element {
		let nodesDom: JSX.Element[] = [];

		this.state.nodes.forEach((node: NodeTypes.NodeRo): void => {
			nodesDom.push(<Node
				key={node.id}
				node={node}
				certificates={this.state.certificates}
				selected={!!this.state.selected[node.id]}
				onSelect={(shift: boolean): void => {
					let selected = {
						...this.state.selected,
					};

					if (shift) {
						let nodes = this.state.nodes;
						let start: number;
						let end: number;

						for (let i = 0; i < nodes.length; i++) {
							let usr = nodes[i];

							if (usr.id === node.id) {
								start = i;
							} else if (usr.id === this.state.lastSelected) {
								end = i;
							}
						}

						if (start !== undefined && end !== undefined) {
							if (start > end) {
								end = [start, start = end][0];
							}

							for (let i = start; i <= end; i++) {
								selected[nodes[i].id] = true;
							}

							this.setState({
								...this.state,
								lastSelected: node.id,
								selected: selected,
							});

							return;
						}
					}

					if (selected[node.id]) {
						delete selected[node.id];
					} else {
						selected[node.id] = true;
					}

					this.setState({
						...this.state,
						lastSelected: node.id,
						selected: selected,
					});
				}}
			/>);
		});

		return <Page>
			<PageHeader>
				<div className="layout horizontal wrap" style={css.header}>
					<h2 style={css.heading}>Nodes</h2>
					<div className="flex"/>
				</div>
			</PageHeader>
			<div style={css.itemsBox}>
				<div style={css.items}>
					{nodesDom}
				</div>
			</div>
			<NodesPage
				onPage={(): void => {
					this.setState({
						lastSelected: null,
					});
				}}
			/>
		</Page>;
	}
}
