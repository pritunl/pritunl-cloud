/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as ImageTypes from '../types/ImageTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import ImagesStore from '../stores/ImagesStore';
import CompletionStore from '../stores/CompletionStore';
import * as ImageActions from '../actions/ImageActions';
import * as CompletionActions from '../actions/CompletionActions';
import Image from './Image';
import ImagesFilter from './ImagesFilter';
import ImagesPage from './ImagesPage';
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
	images: ImageTypes.ImagesRo;
	filter: ImageTypes.Filter;
	organizations: OrganizationTypes.OrganizationsRo;
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

export default class Images extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			images: ImagesStore.images,
			filter: ImagesStore.filter,
			organizations: CompletionStore.organizations,
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
		ImagesStore.addChangeListener(this.onChange);
		CompletionStore.addChangeListener(this.onChange);
		ImageActions.sync();
		CompletionActions.sync();
	}

	componentWillUnmount(): void {
		ImagesStore.removeChangeListener(this.onChange);
		CompletionStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		let images = ImagesStore.images;
		let selected: Selected = {};
		let curSelected = this.state.selected;
		let opened: Opened = {};
		let curOpened = this.state.opened;

		images.forEach((image: ImageTypes.Image): void => {
			if (curSelected[image.id]) {
				selected[image.id] = true;
			}
			if (curOpened[image.id]) {
				opened[image.id] = true;
			}
		});

		this.setState({
			...this.state,
			images: images,
			filter: ImagesStore.filter,
			organizations: CompletionStore.organizations,
			selected: selected,
			opened: opened,
		});
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		ImageActions.removeMulti(
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
		let imagesDom: JSX.Element[] = [];

		this.state.images.forEach((
				image: ImageTypes.ImageRo): void => {
			imagesDom.push(<Image
				key={image.id}
				image={image}
				selected={!!this.state.selected[image.id]}
				open={!!this.state.opened[image.id]}
				onSelect={(shift: boolean): void => {
					let selected = {
						...this.state.selected,
					};

					if (shift) {
						let images = this.state.images;
						let start: number;
						let end: number;

						for (let i = 0; i < images.length; i++) {
							let usr = images[i];

							if (usr.id === image.id) {
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
								selected[images[i].id] = true;
							}

							this.setState({
								...this.state,
								lastSelected: image.id,
								selected: selected,
							});

							return;
						}
					}

					if (selected[image.id]) {
						delete selected[image.id];
					} else {
						selected[image.id] = true;
					}

					this.setState({
						...this.state,
						lastSelected: image.id,
						selected: selected,
					});
				}}
				onOpen={(): void => {
					let opened = {
						...this.state.opened,
					};

					if (opened[image.id]) {
						delete opened[image.id];
					} else {
						opened[image.id] = true;
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
			let inst = ImagesStore.image(instId);
			if (inst) {
				selectedNames.push(inst.name || instId);
			} else {
				selectedNames.push(instId);
			}
		}

		return <Page>
			<PageHeader>
				<div className="layout horizontal wrap" style={css.header}>
					<h2 style={css.heading}>Images</h2>
					<div className="flex"/>
					<div style={css.buttons}>
						<button
							className={filterClass}
							style={css.button}
							type="button"
							onClick={(): void => {
								if (this.state.filter === null) {
									ImageActions.filter({});
								} else {
									ImageActions.filter(null);
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
							confirmMsg="Permanently force delete the selected images"
							confirmInput={true}
							items={selectedNames}
							disabled={!this.selected || this.state.disabled}
							onConfirm={this.onDelete}
						/>
					</div>
				</div>
			</PageHeader>
			<ImagesFilter
				filter={this.state.filter}
				onFilter={(filter): void => {
					ImageActions.filter(filter);
				}}
				organizations={this.state.organizations}
			/>
			<div style={css.itemsBox}>
				<div style={css.items}>
					{imagesDom}
					<tr className="bp5-card bp5-row" style={css.placeholder}>
						<td colSpan={5} style={css.placeholder}/>
					</tr>
				</div>
			</div>
			<NonState
				hidden={!!imagesDom.length}
				iconClass="bp5-icon-compressed"
				title="No images"
				description="Add a new image to get started."
			/>
			<ImagesPage
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
