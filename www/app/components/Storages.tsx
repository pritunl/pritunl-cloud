/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as StorageTypes from '../types/StorageTypes';
import StoragesStore from '../stores/StoragesStore';
import * as StorageActions from '../actions/StorageActions';
import NonState from './NonState';
import Storage from './Storage';
import Page from './Page';
import PageHeader from './PageHeader';

interface State {
	storages: StorageTypes.StoragesRo;
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
		margin: '10px 0 0 0',
	} as React.CSSProperties,
	noCerts: {
		height: 'auto',
	} as React.CSSProperties,
};

export default class Storages extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			storages: StoragesStore.storages,
			disabled: false,
		};
	}

	componentDidMount(): void {
		StoragesStore.addChangeListener(this.onChange);
		StorageActions.sync();
	}

	componentWillUnmount(): void {
		StoragesStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		this.setState({
			...this.state,
			storages: StoragesStore.storages,
		});
	}

	render(): JSX.Element {
		let storagesDom: JSX.Element[] = [];

		this.state.storages.forEach((
				storage: StorageTypes.StorageRo): void => {
			storagesDom.push(<Storage
				key={storage.id}
				storage={storage}
			/>);
		});

		return <Page>
			<PageHeader>
				<div className="layout horizontal wrap" style={css.header}>
					<h2 style={css.heading}>Storages</h2>
					<div className="flex"/>
					<div>
						<button
							className="pt-button pt-intent-success pt-icon-add"
							style={css.button}
							disabled={this.state.disabled}
							type="button"
							onClick={(): void => {
								this.setState({
									...this.state,
									disabled: true,
								});
								StorageActions.create(null).then((): void => {
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
				{storagesDom}
			</div>
			<NonState
				hidden={!!storagesDom.length}
				iconClass="pt-icon-database"
				title="No storages"
				description="Add a new storage to get started."
			/>
		</Page>;
	}
}
