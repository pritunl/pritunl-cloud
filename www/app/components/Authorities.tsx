/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as AuthorityTypes from '../types/AuthorityTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import AuthoritiesStore from '../stores/AuthoritiesStore';
import OrganizationsStore from '../stores/OrganizationsStore';
import * as AuthorityActions from '../actions/AuthorityActions';
import * as OrganizationActions from '../actions/OrganizationActions';
import Authority from './Authority';
import AuthoritiesPage from './AuthoritiesPage';
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
	authorities: AuthorityTypes.AuthoritiesRo;
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

export default class Authorities extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			authorities: AuthoritiesStore.authorities,
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
		AuthoritiesStore.addChangeListener(this.onChange);
		OrganizationsStore.addChangeListener(this.onChange);
		AuthorityActions.sync();
		OrganizationActions.sync();
	}

	componentWillUnmount(): void {
		AuthoritiesStore.removeChangeListener(this.onChange);
		OrganizationsStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		let authorities = AuthoritiesStore.authorities;
		let selected: Selected = {};
		let curSelected = this.state.selected;
		let opened: Opened = {};
		let curOpened = this.state.opened;

		authorities.forEach((authority: AuthorityTypes.Authority): void => {
			if (curSelected[authority.id]) {
				selected[authority.id] = true;
			}
			if (curOpened[authority.id]) {
				opened[authority.id] = true;
			}
		});

		this.setState({
			...this.state,
			authorities: authorities,
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
		AuthorityActions.removeMulti(
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
		let authoritiesDom: JSX.Element[] = [];

		this.state.authorities.forEach((
				authority: AuthorityTypes.AuthorityRo): void => {
			authoritiesDom.push(<Authority
				key={authority.id}
				authority={authority}
				organizations={this.state.organizations}
				selected={!!this.state.selected[authority.id]}
				open={!!this.state.opened[authority.id]}
				onSelect={(shift: boolean): void => {
					let selected = {
						...this.state.selected,
					};

					if (shift) {
						let authorities = this.state.authorities;
						let start: number;
						let end: number;

						for (let i = 0; i < authorities.length; i++) {
							let usr = authorities[i];

							if (usr.id === authority.id) {
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
								selected[authorities[i].id] = true;
							}

							this.setState({
								...this.state,
								lastSelected: authority.id,
								selected: selected,
							});

							return;
						}
					}

					if (selected[authority.id]) {
						delete selected[authority.id];
					} else {
						selected[authority.id] = true;
					}

					this.setState({
						...this.state,
						lastSelected: authority.id,
						selected: selected,
					});
				}}
				onOpen={(): void => {
					let opened = {
						...this.state.opened,
					};

					if (opened[authority.id]) {
						delete opened[authority.id];
					} else {
						opened[authority.id] = true;
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
					<h2 style={css.heading}>Authorities</h2>
					<div className="flex"/>
					<div style={css.buttons}>
						<button
							className="pt-button pt-intent-warning pt-icon-chevron-up"
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
							className="pt-intent-danger pt-icon-delete"
							progressClassName="pt-intent-danger"
							style={css.button}
							disabled={!this.selected || this.state.disabled}
							onConfirm={this.onDelete}
						/>
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
								AuthorityActions.create({
									name: 'New Authority',
								} as AuthorityTypes.Authority).then((): void => {
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
			<div style={css.itemsBox}>
				<div style={css.items}>
					{authoritiesDom}
					<tr className="pt-card pt-row" style={css.placeholder}>
						<td colSpan={5} style={css.placeholder}/>
					</tr>
				</div>
			</div>
			<NonState
				hidden={!!authoritiesDom.length}
				iconClass="pt-icon-office"
				title="No authorities"
				description="Add a new authority to get started."
			/>
			<AuthoritiesPage
				onPage={(): void => {
					this.setState({
						lastSelected: null,
					});
				}}
			/>
		</Page>;
	}
}
