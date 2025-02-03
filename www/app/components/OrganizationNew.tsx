/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as OrganizationTypes from '../types/OrganizationTypes';
import * as OrganizationActions from '../actions/OrganizationActions';
import PageInput from './PageInput';
import PageCreate from './PageCreate';
import PageInputButton from './PageInputButton';
import Help from './Help';
import PageTextArea from "./PageTextArea";

interface Props {
	onClose: () => void;
}

interface State {
	closed: boolean;
	disabled: boolean;
	changed: boolean;
	message: string;
	addRole: string;
	organization: OrganizationTypes.Organization;
}

const css = {
	row: {
		display: 'table-row',
		width: '100%',
		padding: 0,
		boxShadow: 'none',
		position: 'relative',
	} as React.CSSProperties,
	card: {
		position: 'relative',
		padding: '10px 10px 0 10px',
		width: '100%',
	} as React.CSSProperties,
	remove: {
		position: 'absolute',
		top: '5px',
		right: '5px',
	} as React.CSSProperties,
	role: {
		margin: '9px 5px 0 5px',
		height: '20px',
	} as React.CSSProperties,
	group: {
		flex: 1,
		minWidth: '280px',
		margin: '0 10px',
	} as React.CSSProperties,
	save: {
		paddingBottom: '10px',
	} as React.CSSProperties,
	button: {
		height: '30px',
	} as React.CSSProperties,
	buttons: {
		position: 'absolute',
		top: '5px',
		right: '5px',
	} as React.CSSProperties,
	select: {
		margin: '7px 0px 0px 6px',
		paddingTop: '3px',
	} as React.CSSProperties,
};

export default class Organization extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			closed: false,
			disabled: false,
			changed: false,
			message: '',
			addRole: '',
			organization: {
				name: 'New Organization',
			},
		};
	}

	set(name: string, val: any): void {
		let organization: any = {
			...this.state.organization,
		};

		organization[name] = val;

		this.setState({
			...this.state,
			changed: true,
			organization: organization,
		});
	}

	onAddRole = (): void => {
		let organization: OrganizationTypes.Organization;

		organization = {
			...this.state.organization,
		};

		let roles = [
			...organization.roles,
		];

		if (!this.state.addRole) {
			return;
		}

		if (roles.indexOf(this.state.addRole) === -1) {
			roles.push(this.state.addRole);
		}

		roles.sort();

		organization.roles = roles;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addRole: '',
			organization: organization,
		});
	}

	onRemoveRole(role: string): void {
		let organization: OrganizationTypes.Organization;

		organization = {
			...this.state.organization,
		};

		let roles = [
			...organization.roles,
		];

		let i = roles.indexOf(role);
		if (i === -1) {
			return;
		}

		roles.splice(i, 1);

		organization.roles = roles;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addRole: '',
			organization: organization,
		});
	}

	onCreate = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});

		let organization: any = {
			...this.state.organization,
		};

		OrganizationActions.create(organization).then((): void => {
			this.setState({
				...this.state,
				message: 'Organization created successfully',
				changed: false,
			});

			setTimeout((): void => {
				this.setState({
					...this.state,
					disabled: false,
					changed: true,
				});
			}, 2000);
		}).catch((): void => {
			this.setState({
				...this.state,
				message: '',
				disabled: false,
			});
		});
	}

	render(): JSX.Element {
		let org: OrganizationTypes.Organization = this.state.organization;

		let roles: JSX.Element[] = [];
		for (let role of (org.roles || [])) {
			roles.push(
				<div
					className="bp5-tag bp5-tag-removable bp5-intent-primary"
					style={css.role}
					key={role}
				>
					{role}
					<button
						className="bp5-tag-remove"
						disabled={this.state.disabled}
						onMouseUp={(): void => {
							this.onRemoveRole(role);
						}}
					/>
				</div>,
			);
		}

		return <div
		className="bp5-card bp5-row"
		style={css.row}
	>
		<td
			className="bp5-cell"
			colSpan={2}
			style={css.card}
		>
			<div className="layout horizontal wrap">
				<div style={css.group}>
					<div style={css.buttons}>
					</div>
						<PageInput
							label="Name"
							help="Name of organization"
							type="text"
							placeholder="Name"
							value={org.name}
							onChange={(val): void => {
								this.set('name', val);
							}}
						/>
						<PageTextArea
							label="Comment"
							help="Organization comment."
							placeholder="Organization comment"
							rows={3}
							value={org.comment}
							onChange={(val: string): void => {
								this.set('comment', val);
							}}
						/>
					</div>
					<div style={css.group}>
						<label className="bp5-label">
							Roles
							<Help
								title="Roles"
								content="User roles will be used to match with organization roles. A user must have a matching role to access an organization."
							/>
							<div>
								{roles}
							</div>
						</label>
						<PageInputButton
							disabled={this.state.disabled}
							buttonClass="bp5-intent-success bp5-icon-add"
							label="Add"
							type="text"
							placeholder="Add role"
							value={this.state.addRole}
							onChange={(val): void => {
								this.setState({
									...this.state,
									addRole: val,
								});
							}}
							onSubmit={this.onAddRole}
						/>
					</div>
				</div>
				<PageCreate
					style={css.save}
					hidden={!this.state.organization}
					message={this.state.message}
					changed={this.state.changed}
					disabled={this.state.disabled}
					closed={this.state.closed}
					light={true}
					onCancel={this.props.onClose}
					onCreate={this.onCreate}
				/>
			</td>
		</div>;
	}
}
