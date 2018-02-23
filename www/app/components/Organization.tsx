/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as OrganizationTypes from '../types/OrganizationTypes';
import * as OrganizationActions from '../actions/OrganizationActions';
import PageInput from './PageInput';
import PageInfo from './PageInfo';
import PageSave from './PageSave';
import PageInputButton from './PageInputButton';
import ConfirmButton from './ConfirmButton';
import Help from './Help';

interface Props {
	organization: OrganizationTypes.OrganizationRo;
}

interface State {
	disabled: boolean;
	changed: boolean;
	message: string;
	addRole: string;
	organization: OrganizationTypes.Organization;
}

const css = {
	card: {
		position: 'relative',
		padding: '10px 10px 0 10px',
		marginBottom: '5px',
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
		minWidth: '250px',
	} as React.CSSProperties,
	save: {
		paddingBottom: '10px',
	} as React.CSSProperties,
};

export default class Organization extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			changed: false,
			message: '',
			addRole: '',
			organization: null,
		};
	}

	set(name: string, val: any): void {
		let organization: any;

		if (this.state.changed) {
			organization = {
				...this.state.organization,
			};
		} else {
			organization = {
				...this.props.organization,
			};
		}

		organization[name] = val;

		this.setState({
			...this.state,
			changed: true,
			organization: organization,
		});
	}

	onAddRole = (): void => {
		let organization: OrganizationTypes.Organization;

		if (this.state.changed) {
			organization = {
				...this.state.organization,
			};
		} else {
			organization = {
				...this.props.organization,
			};
		}

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

		if (this.state.changed) {
			organization = {
				...this.state.organization,
			};
		} else {
			organization = {
				...this.props.organization,
			};
		}

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

	onSave = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		OrganizationActions.commit(this.state.organization).then((): void => {
			this.setState({
				...this.state,
				message: 'Your changes have been saved',
				changed: false,
				disabled: false,
			});

			setTimeout((): void => {
				if (!this.state.changed) {
					this.setState({
						...this.state,
						message: '',
						changed: false,
						organization: null,
					});
				}
			}, 3000);
		}).catch((): void => {
			this.setState({
				...this.state,
				message: '',
				disabled: false,
			});
		});
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		OrganizationActions.remove(this.props.organization.id).then((): void => {
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
	}

	render(): JSX.Element {
		let org: OrganizationTypes.Organization = this.state.organization ||
			this.props.organization;

		let roles: JSX.Element[] = [];
		for (let role of (org.roles || [])) {
			roles.push(
				<div
					className="pt-tag pt-tag-removable pt-intent-primary"
					style={css.role}
					key={role}
				>
					{role}
					<button
						className="pt-tag-remove"
						disabled={this.state.disabled}
						onMouseUp={(): void => {
							this.onRemoveRole(role);
						}}
					/>
				</div>,
			);
		}

		return <div
			className="pt-card"
			style={css.card}
		>
			<div className="layout horizontal wrap">
				<div style={css.group}>
					<div style={css.remove}>
						<ConfirmButton
							className="pt-minimal pt-intent-danger pt-icon-cross"
							progressClassName="pt-intent-danger"
							confirmMsg="Confirm organization remove"
							disabled={this.state.disabled}
							onConfirm={this.onDelete}
						/>
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
					<label className="pt-label">
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
						buttonClass="pt-intent-success pt-icon-add"
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
				<div style={css.group}>
					<PageInfo
						fields={[
							{
								label: 'ID',
								value: this.props.organization.id || 'None',
							},
						]}
					/>
				</div>
			</div>
			<PageSave
				style={css.save}
				hidden={!this.state.organization}
				message={this.state.message}
				changed={this.state.changed}
				disabled={this.state.disabled}
				light={true}
				onCancel={(): void => {
					this.setState({
						...this.state,
						changed: false,
						organization: null,
					});
				}}
				onSave={this.onSave}
			/>
		</div>;
	}
}
