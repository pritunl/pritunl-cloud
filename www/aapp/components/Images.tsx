/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as ImageTypes from '../types/ImageTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import ImagesStore from '../stores/ImagesStore';
import OrganizationsStore from '../stores/OrganizationsStore';
import * as ImageActions from '../actions/ImageActions';
import * as OrganizationActions from '../actions/OrganizationActions';
import Image from './Image';
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
		margin: '10px 0 0 10px',
	} as React.CSSProperties,
};

export default class Images extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			images: ImagesStore.images,
			organizations: OrganizationsStore.organizations,
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
		OrganizationsStore.addChangeListener(this.onChange);
		ImageActions.sync();
		OrganizationActions.sync();
	}

	componentWillUnmount(): void {
		ImagesStore.removeChangeListener(this.onChange);
		OrganizationsStore.removeChangeListener(this.onChange);
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
			organizations: OrganizationsStore.organizations,
			selected: selected,
			opened: opened,
		});
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		// ImageActions.removeMulti(
		// 		Object.keys(this.state.selected)).then((): void => {
		// 	this.setState({
		// 		...this.state,
		// 		selected: {},
		// 		disabled: false,
		// 	});
		// }).catch((): void => {
		// 	this.setState({
		// 		...this.state,
		// 		disabled: false,
		// 	});
		// });
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

		return <Page>
			<PageHeader>
				<div className="layout horizontal wrap" style={css.header}>
					<h2 style={css.heading}>Images</h2>
					<div className="flex"/>
					<div>
						<button
							className="pt-button pt-intent-warning pt-icon-chevron-up"
							style={css.button}
							hidden={!this.opened}
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
							className="pt-intent-danger pt-icon-delete"
							progressClassName="pt-intent-danger"
							style={css.button}
							disabled={!this.selected || this.state.disabled}
							onConfirm={this.onDelete}
						/>
					</div>
				</div>
			</PageHeader>
			<div style={css.itemsBox}>
				<div style={css.items}>
					{imagesDom}
					<tr className="pt-card pt-row" style={css.placeholder}>
						<td colSpan={5} style={css.placeholder}/>
					</tr>
				</div>
			</div>
			<NonState
				hidden={!!imagesDom.length}
				iconClass="pt-icon-floppy-disk"
				title="No images"
				description="Add a new image to get started."
			/>
			<ImagesPage
				onPage={(): void => {
					this.setState({
						lastSelected: null,
					});
				}}
			/>
		</Page>;
	}
}
