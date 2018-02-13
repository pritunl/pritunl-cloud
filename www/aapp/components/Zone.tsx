/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as ZoneTypes from '../types/ZoneTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import * as ZoneActions from '../actions/ZoneActions';
import OrganizationsStore from '../stores/OrganizationsStore';
import PageInput from './PageInput';
import PageSelectButton from './PageSelectButton';
import PageInfo from './PageInfo';
import PageSave from './PageSave';
import ConfirmButton from './ConfirmButton';
import Help from './Help';

interface Props {
	zone: ZoneTypes.ZoneRo;
	organizations: OrganizationTypes.OrganizationsRo;
}

interface State {
	disabled: boolean;
	changed: boolean;
	message: string;
	zone: ZoneTypes.Zone;
	addOrganization: string;
	addCert: string;
	forwardedChecked: boolean;
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
	item: {
		margin: '9px 5px 0 5px',
		height: '20px',
	} as React.CSSProperties,
	itemsLabel: {
		display: 'block',
	} as React.CSSProperties,
	itemsAdd: {
		margin: '8px 0 15px 0',
	} as React.CSSProperties,
	group: {
		flex: 1,
		minWidth: '250px',
	} as React.CSSProperties,
	save: {
		paddingBottom: '10px',
	} as React.CSSProperties,
	label: {
		width: '100%',
		maxWidth: '280px',
	} as React.CSSProperties,
	inputGroup: {
		width: '100%',
	} as React.CSSProperties,
	protocol: {
		flex: '0 1 auto',
	} as React.CSSProperties,
	port: {
		flex: '1',
	} as React.CSSProperties,
};

export default class Zone extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			changed: false,
			message: '',
			zone: null,
			addOrganization: null,
			addCert: null,
			forwardedChecked: false,
		};
	}

	set(name: string, val: any): void {
		let zone: any;

		if (this.state.changed) {
			zone = {
				...this.state.zone,
			};
		} else {
			zone = {
				...this.props.zone,
			};
		}

		zone[name] = val;

		this.setState({
			...this.state,
			changed: true,
			zone: zone,
		});
	}

	onSave = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		ZoneActions.commit(this.state.zone).then((): void => {
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
						zone: null,
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
		ZoneActions.remove(this.props.zone.id).then((): void => {
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

	onAddOrganization = (): void => {
		let zone: ZoneTypes.Zone;

		if (!this.state.addOrganization && !this.props.organizations.length) {
			return;
		}

		let organizationId = this.state.addOrganization ||
			this.props.organizations[0].id;

		if (this.state.changed) {
			zone = {
				...this.state.zone,
			};
		} else {
			zone = {
				...this.props.zone,
			};
		}

		let organizations = [
			...(zone.organizations || []),
		];

		if (organizations.indexOf(organizationId) === -1) {
			organizations.push(organizationId);
		}

		organizations.sort();

		zone.organizations = organizations;

		this.setState({
			...this.state,
			changed: true,
			zone: zone,
		});
	}

	onRemoveOrganization = (organization: string): void => {
		let zone: ZoneTypes.Zone;

		if (this.state.changed) {
			zone = {
				...this.state.zone,
			};
		} else {
			zone = {
				...this.props.zone,
			};
		}

		let organizations = [
			...(zone.organizations || []),
		];

		let i = organizations.indexOf(organization);
		if (i === -1) {
			return;
		}

		organizations.splice(i, 1);

		zone.organizations = organizations;

		this.setState({
			...this.state,
			changed: true,
			zone: zone,
		});
	}

	render(): JSX.Element {
		let zone: ZoneTypes.Zone = this.state.zone ||
			this.props.zone;

		let organizations: JSX.Element[] = [];
		for (let organizationId of (zone.organizations || [])) {
			let organization = OrganizationsStore.organization(organizationId);
			if (!organization) {
				continue;
			}

			organizations.push(
				<div
					className="pt-tag pt-tag-removable pt-intent-primary"
					style={css.item}
					key={organization.id}
				>
					{organization.name}
					<button
						className="pt-tag-remove"
						onMouseUp={(): void => {
							this.onRemoveOrganization(organization.id);
						}}
					/>
				</div>,
			);
		}

		let organizationsSelect: JSX.Element[] = [];
		if (this.props.organizations.length) {
			for (let organization of this.props.organizations) {
				organizationsSelect.push(
					<option
						key={organization.id}
						value={organization.id}
					>{organization.name}</option>,
				);
			}
		} else {
			organizationsSelect.push(<option key="null" value="">None</option>);
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
							confirmMsg="Confirm zone remove"
							disabled={this.state.disabled}
							onConfirm={this.onDelete}
						/>
					</div>
					<PageInput
						label="Name"
						help="Name of zone"
						type="text"
						placeholder="Enter name"
						value={zone.name}
						onChange={(val): void => {
							this.set('name', val);
						}}
					/>
					<label
						className="pt-label"
						style={css.label}
					>
						Organizations
						<Help
							title="Organizations"
							content="Organizations that can access this zone."
						/>
						<div>
							{organizations}
						</div>
					</label>
					<PageSelectButton
						label="Add Organization"
						value={this.state.addOrganization}
						disabled={!this.props.organizations.length}
						buttonClass="pt-intent-success"
						onChange={(val: string): void => {
							this.setState({
								...this.state,
								addOrganization: val,
							});
						}}
						onSubmit={this.onAddOrganization}
					>
						{organizationsSelect}
					</PageSelectButton>
				</div>
				<div style={css.group}>
					<PageInfo
						fields={[
							{
								label: 'ID',
								value: zone.id || 'None',
							},
							{
								label: 'Datacenter',
								value: zone.datacenter || 'None',
							},
						]}
					/>
				</div>
			</div>
			<PageSave
				style={css.save}
				hidden={!this.state.zone}
				message={this.state.message}
				changed={this.state.changed}
				disabled={this.state.disabled}
				light={true}
				onCancel={(): void => {
					this.setState({
						...this.state,
						changed: false,
						forwardedChecked: false,
						zone: null,
					});
				}}
				onSave={this.onSave}
			/>
		</div>;
	}
}
