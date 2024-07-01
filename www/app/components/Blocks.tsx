/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as BlockTypes from '../types/BlockTypes';
import BlocksStore from '../stores/BlocksStore';
import * as BlockActions from '../actions/BlockActions';
import NonState from './NonState';
import Block from './Block';
import Page from './Page';
import PageHeader from './PageHeader';

interface State {
	blocks: BlockTypes.BlocksRo;
	disabled: boolean;
}

const css = {
	header: {
		marginTop: '-19px',
	} as React.CSSProperties,
	heading: {
		margin: '19px 0 0 0',
	} as React.CSSProperties,
	button: {
		margin: '8px 0 0 8px',
	} as React.CSSProperties,
	buttons: {
		marginTop: '8px',
	} as React.CSSProperties,
	noCerts: {
		height: 'auto',
	} as React.CSSProperties,
};

export default class Blocks extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			blocks: BlocksStore.blocks,
			disabled: false,
		};
	}

	componentDidMount(): void {
		BlocksStore.addChangeListener(this.onChange);
		BlockActions.sync();
	}

	componentWillUnmount(): void {
		BlocksStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		this.setState({
			...this.state,
			blocks: BlocksStore.blocks,
		});
	}

	render(): JSX.Element {
		let blocksDom: JSX.Element[] = [];

		this.state.blocks.forEach((
				block: BlockTypes.BlockRo): void => {
			blocksDom.push(<Block
				key={block.id}
				block={block}
			/>);
		});

		return <Page>
			<PageHeader>
				<div className="layout horizontal wrap" style={css.header}>
					<h2 style={css.heading}>Blocks</h2>
					<div className="flex"/>
					<div style={css.buttons}>
						<button
							className="bp5-button bp5-intent-success bp5-icon-add"
							style={css.button}
							disabled={this.state.disabled}
							type="button"
							onClick={(): void => {
								this.setState({
									...this.state,
									disabled: true,
								});
								BlockActions.create(null).then((): void => {
									this.setState({
										...this.state,
										disabled: false,
									});
								}).catch((): void => {
									this.setState({
										...this.state,
										disabled: false,
									});
								});
							}}
						>New</button>
					</div>
				</div>
			</PageHeader>
			<div>
				{blocksDom}
			</div>
			<NonState
				hidden={!!blocksDom.length}
				iconClass="bp5-icon-ip-address"
				title="No IP blocks"
				description="Add a new IP block to get started."
			/>
		</Page>;
	}
}
