/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as StorageTypes from '../types/StorageTypes';
import StoragesStore from '../stores/StoragesStore';
import * as StorageActions from '../actions/StorageActions';
import Storage from './Storage';
import StorageNew from './StorageNew';
import StoragesFilter from './StoragesFilter';
import StoragesPage from './StoragesPage';
import Page from './Page';
import PageHeader from './PageHeader';
import NonState from './NonState';
import ConfirmButton from './ConfirmButton';

interface Selected {
	[key: string]: boolean;
}

interface Opened {
	[key: string]: boolean;
}

interface State {
	storages: StorageTypes.StoragesRo;
	filter: StorageTypes.Filter;
	selected: Selected;
	opened: Opened;
	newOpened: boolean;
	lastSelected: string;
	disabled: boolean;
}

const css = {
	items: {
		width: '100%',
		marginTop: '-5px',
		display: 'table',
		tableLayout: 'fixed',
		borderSpacing: '0 5px',
	} as React.CSSProperties,
	itemsBox: {
		width: '100%',
		overflowY: 'auto',
	} as React.CSSProperties,
	placeholder: {
		opacity: 0,
		width: '100%',
	} as React.CSSProperties,
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
};

export default class Storages extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			storages: StoragesStore.storages,
			filter: StoragesStore.filter,
			selected: {},
			opened: {},
			newOpened: false,
			lastSelected: null,
			disabled: false,
		};
	}

	get selected(): boolean {
		return !!Object.keys(this.state.selected).length;
	}

	get opened(): boolean {
		return !!Object.keys(this.state.opened).length;
	}

	componentDidMount(): void {
		StoragesStore.addChangeListener(this.onChange);
		StorageActions.sync();
	}

	componentWillUnmount(): void {
		StoragesStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		let storages = StoragesStore.storages;
		let selected: Selected = {};
		let curSelected = this.state.selected;
		let opened: Opened = {};
		let curOpened = this.state.opened;

		storages.forEach((storage: StorageTypes.Storage): void => {
			if (curSelected[storage.id]) {
				selected[storage.id] = true;
			}
			if (curOpened[storage.id]) {
				opened[storage.id] = true;
			}
		});

		this.setState({
			...this.state,
			storages: storages,
			filter: StoragesStore.filter,
			selected: selected,
			opened: opened,
		});
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		StorageActions.removeMulti(
				Object.keys(this.state.selected)).then((): void => {
			this.setState({
				...this.state,
				selected: {},
				disabled: false,
			});
		}).catch((): void => {
			this.setState({
				...this.state,
				disabled: false,
			});
		});
	}

	render(): JSX.Element {
		let storagesDom: JSX.Element[] = [];

		this.state.storages.forEach((
				storage: StorageTypes.StorageRo): void => {
			storagesDom.push(<Storage
				key={storage.id}
				storage={storage}
				selected={!!this.state.selected[storage.id]}
				open={!!this.state.opened[storage.id]}
				onSelect={(shift: boolean): void => {
					let selected = {
						...this.state.selected,
					};

					if (shift) {
						let storages = this.state.storages;
						let start: number;
						let end: number;

						for (let i = 0; i < storages.length; i++) {
							let usr = storages[i];

							if (usr.id === storage.id) {
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
								selected[storages[i].id] = true;
							}

							this.setState({
								...this.state,
								lastSelected: storage.id,
								selected: selected,
							});

							return;
						}
					}

					if (selected[storage.id]) {
						delete selected[storage.id];
					} else {
						selected[storage.id] = true;
					}

					this.setState({
						...this.state,
						lastSelected: storage.id,
						selected: selected,
					});
				}}
				onOpen={(): void => {
					let opened = {
						...this.state.opened,
					};

					if (opened[storage.id]) {
						delete opened[storage.id];
					} else {
						opened[storage.id] = true;
					}

					this.setState({
						...this.state,
						opened: opened,
					});
				}}
			/>);
		});

		let filterClass = 'bp5-button bp5-intent-primary bp5-icon-filter ';
		if (this.state.filter) {
			filterClass += 'bp5-active';
		}

		let selectedNames: string[] = [];
		for (let instId of Object.keys(this.state.selected)) {
			let inst = StoragesStore.storage(instId);
			if (inst) {
				selectedNames.push(inst.name || instId);
			} else {
				selectedNames.push(instId);
			}
		}

		let newStorageDom: JSX.Element;
		if (this.state.newOpened) {
			newStorageDom = <StorageNew
				onClose={(): void => {
					this.setState({
						...this.state,
						newOpened: false,
					});
				}}
			/>;
		}

		return <Page>
			<PageHeader>
				<div className="layout horizontal wrap" style={css.header}>
					<h2 style={css.heading}>Storages</h2>
					<div className="flex"/>
					<div style={css.buttons}>
						<button
							className={filterClass}
							style={css.button}
							type="button"
							onClick={(): void => {
								if (this.state.filter === null) {
									StorageActions.filter({});
								} else {
									StorageActions.filter(null);
								}
							}}
						>
							Filters
						</button>
						<button
							className="bp5-button bp5-intent-warning bp5-icon-chevron-up"
							style={css.button}
							disabled={!this.opened}
							type="button"
							onClick={(): void => {
								this.setState({
									...this.state,
									opened: {},
								});
							}}
						>
							Collapse All
						</button>
						<ConfirmButton
							label="Delete Selected"
							className="bp5-intent-danger bp5-icon-delete"
							progressClassName="bp5-intent-danger"
							safe={true}
							style={css.button}
							confirmMsg="Permanently delete the selected storages"
							confirmInput={true}
							items={selectedNames}
							disabled={!this.selected || this.state.disabled}
							onConfirm={this.onDelete}
						/>
						<button
							className="bp5-button bp5-intent-success bp5-icon-add"
							style={css.button}
							disabled={this.state.disabled || this.state.newOpened}
							type="button"
							onClick={(): void => {
								this.setState({
									...this.state,
									newOpened: true,
								});
							}}
						>New</button>
					</div>
				</div>
			</PageHeader>
			<StoragesFilter
				filter={this.state.filter}
				onFilter={(filter): void => {
					StorageActions.filter(filter);
				}}
			/>
			<div style={css.itemsBox}>
				<div style={css.items}>
					{newStorageDom}
					{storagesDom}
					<tr className="bp5-card bp5-row" style={css.placeholder}>
						<td colSpan={2} style={css.placeholder}/>
					</tr>
				</div>
			</div>
			<NonState
				hidden={!!storagesDom.length}
				iconClass="bp5-icon-database"
				title="No storages"
				description="Add a new storage to get started."
			/>
			<StoragesPage
				onPage={(): void => {
					this.setState({
						...this.state,
						selected: {},
						lastSelected: null,
					});
				}}
			/>
		</Page>;
	}
}
