/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as DatacenterTypes from '../types/DatacenterTypes';
import * as StorageTypes from '../types/StorageTypes';
import * as DatacenterActions from '../actions/DatacenterActions';
import * as OrganizationTypes from "../types/OrganizationTypes";
import CompletionStore from '../stores/CompletionStore';
import PageInput from './PageInput';
import PageInfo from './PageInfo';
import PageSelect from './PageSelect';
import PageSelectButton from './PageSelectButton';
import PageSwitch from './PageSwitch';
import PageSave from './PageSave';
import ConfirmButton from './ConfirmButton';
import Help from './Help';
import PageTextArea from "./PageTextArea";

interface Props {
	datacenter: DatacenterTypes.DatacenterRo;
	organizations: OrganizationTypes.OrganizationsRo;
	storages: StorageTypes.StoragesRo;
	selected: boolean;
	onSelect: (shift: boolean) => void;
	onClose: () => void;
}

interface State {
	disabled: boolean;
	changed: boolean;
	message: string;
	datacenter: DatacenterTypes.Datacenter;
	addStorage: string;
	addOrganization: string;
}

const css = {
	card: {
		position: 'relative',
		padding: '48px 10px 0 10px',
		width: '100%',
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
		minWidth: '280px',
		margin: '0 10px',
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
	button: {
		height: '30px',
	} as React.CSSProperties,
	buttons: {
		cursor: 'pointer',
		position: 'absolute',
		top: 0,
		left: 0,
		right: 0,
		padding: '4px',
		height: '39px',
	} as React.CSSProperties,
	select: {
		margin: '7px 0px 0px 6px',
		paddingTop: '3px',
	} as React.CSSProperties,
};

export default class DatacenterDetailed extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			changed: false,
			message: '',
			datacenter: null,
			addStorage: '',
			addOrganization: null,
		};
	}

	set(name: string, val: any): void {
		let datacenter: any;

		if (this.state.changed) {
			datacenter = {
				...this.state.datacenter,
			};
		} else {
			datacenter = {
				...this.props.datacenter,
			};
		}

		datacenter[name] = val;

		this.setState({
			...this.state,
			changed: true,
			datacenter: datacenter,
		});
	}

	toggle(name: string): void {
		let datacenter: any;

		if (this.state.changed) {
			datacenter = {
				...this.state.datacenter,
			};
		} else {
			datacenter = {
				...this.props.datacenter,
			};
		}

		datacenter[name] = !datacenter[name];

		this.setState({
			...this.state,
			changed: true,
			datacenter: datacenter,
		});
	}

	onSave = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		DatacenterActions.commit(this.state.datacenter).then((): void => {
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
						datacenter: null,
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
		DatacenterActions.remove(this.props.datacenter.id).then((): void => {
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

	onAddStorage = (): void => {
		let datacenter: DatacenterTypes.Datacenter;

		if (!this.state.addStorage && !this.props.storages.length) {
			return;
		}

		let storageId = this.state.addStorage;
		if (!storageId) {
			for (let store of this.props.storages) {
				if (store.type === "public" || store.type === "web") {
					storageId = store.id
					break
				}
			}
		}

		if (this.state.changed) {
			datacenter = {
				...this.state.datacenter,
			};
		} else {
			datacenter = {
				...this.props.datacenter,
			};
		}

		let publicStorages = [
			...(datacenter.public_storages || []),
		];

		if (publicStorages.indexOf(storageId) === -1) {
			publicStorages.push(storageId);
		}

		publicStorages.sort();

		datacenter.public_storages = publicStorages;

		this.setState({
			...this.state,
			changed: true,
			datacenter: datacenter,
		});
	}

	onRemoveStorage = (storage: string): void => {
		let datacenter: DatacenterTypes.Datacenter;

		if (this.state.changed) {
			datacenter = {
				...this.state.datacenter,
			};
		} else {
			datacenter = {
				...this.props.datacenter,
			};
		}

		let publicStorages = [
			...(datacenter.public_storages || []),
		];

		let i = publicStorages.indexOf(storage);
		if (i === -1) {
			return;
		}

		publicStorages.splice(i, 1);

		datacenter.public_storages = publicStorages;

		this.setState({
			...this.state,
			changed: true,
			datacenter: datacenter,
		});
	}

	onAddOrganization = (): void => {
		let datacenter: DatacenterTypes.Datacenter;

		if (!this.state.addOrganization && !this.props.organizations.length) {
			return;
		}

		let organizationId = this.state.addOrganization ||
			this.props.organizations[0].id;

		if (this.state.changed) {
			datacenter = {
				...this.state.datacenter,
			};
		} else {
			datacenter = {
				...this.props.datacenter,
			};
		}

		let organizations = [
			...(datacenter.organizations || []),
		];

		if (organizations.indexOf(organizationId) === -1) {
			organizations.push(organizationId);
		}

		organizations.sort();

		datacenter.organizations = organizations;

		this.setState({
			...this.state,
			changed: true,
			datacenter: datacenter,
		});
	}

	onRemoveOrganization = (organization: string): void => {
		let datacenter: DatacenterTypes.Datacenter;

		if (this.state.changed) {
			datacenter = {
				...this.state.datacenter,
			};
		} else {
			datacenter = {
				...this.props.datacenter,
			};
		}

		let organizations = [
			...(datacenter.organizations || []),
		];

		let i = organizations.indexOf(organization);
		if (i === -1) {
			return;
		}

		organizations.splice(i, 1);

		datacenter.organizations = organizations;

		this.setState({
			...this.state,
			changed: true,
			datacenter: datacenter,
		});
	}

	render(): JSX.Element {
		let datacenter: DatacenterTypes.Datacenter = this.state.datacenter ||
			this.props.datacenter;

		let organizations: JSX.Element[] = [];
		for (let organizationId of (datacenter.organizations || [])) {
			let organization = CompletionStore.organization(organizationId);
			if (!organization) {
				continue;
			}

			organizations.push(
				<div
					className="bp5-tag bp5-tag-removable bp5-intent-primary"
					style={css.item}
					key={organization.id}
				>
					{organization.name}
					<button
						className="bp5-tag-remove"
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

		let publicStorages: JSX.Element[] = [];
		for (let storageId of (datacenter.public_storages || [])) {
			let storage = CompletionStore.storage(storageId);
			if (!storage) {
				continue;
			}

			publicStorages.push(
				<div
					className="bp5-tag bp5-tag-removable bp5-intent-primary"
					style={css.item}
					key={storage.id}
				>
					{storage.name}
					<button
						disabled={this.state.disabled}
						className="bp5-tag-remove"
						onMouseUp={(): void => {
							this.onRemoveStorage(storage.id);
						}}
					/>
				</div>,
			);
		}

		let hasStorages = false;
		let privateStoragesSelect: JSX.Element[] = [
			<option key="null" value="">None</option>,
		];
		let backupStoragesSelect: JSX.Element[] = [
			<option key="null" value="">None</option>,
		];
		let publicStoragesSelect: JSX.Element[] = [];
		if (this.props.storages.length) {
			for (let storage of this.props.storages) {
				if (storage.type === 'public' || storage.type === 'web') {
					hasStorages = true;
					publicStoragesSelect.push(
						<option
							key={storage.id}
							value={storage.id}
						>{storage.name}</option>,
					);
				} else if (storage.type === 'private') {
					privateStoragesSelect.push(
						<option
							key={storage.id}
							value={storage.id}
						>{storage.name}</option>,
					);
					backupStoragesSelect.push(
						<option
							key={storage.id}
							value={storage.id}
						>{storage.name}</option>,
					);
				}
			}
		}

		if (!hasStorages) {
			publicStoragesSelect.push(
				<option key="null" value="">No Storages</option>);
		}

		return <td
			className="bp5-cell"
			colSpan={2}
			style={css.card}
		>
			<div className="layout horizontal wrap">
				<div style={css.group}>
					<div
						className="layout horizontal tab-close bp5-card-header"
						style={css.buttons}
						onClick={(evt): void => {
							if (evt.target instanceof HTMLElement &&
									evt.target.className.indexOf('tab-close') !== -1) {
								this.props.onClose();
							}
						}}
					>
						<div>
							<label
								className="bp5-control bp5-checkbox"
								style={css.select}
							>
								<input
									type="checkbox"
									checked={this.props.selected}
									onChange={(evt): void => {
									}}
									onClick={(evt): void => {
										this.props.onSelect(evt.shiftKey);
									}}
								/>
								<span className="bp5-control-indicator"/>
							</label>
						</div>
						<div className="flex tab-close"/>
						<ConfirmButton
							className="bp5-minimal bp5-intent-danger bp5-icon-trash"
							style={css.button}
							safe={true}
							progressClassName="bp5-intent-danger"
							dialogClassName="bp5-intent-danger bp5-icon-delete"
							dialogLabel="Delete Datacenter"
							confirmMsg="Permanently delete this datacenter"
							confirmInput={true}
							items={[datacenter.name]}
							disabled={this.state.disabled}
							onConfirm={this.onDelete}
						/>
					</div>
					<PageInput
						disabled={this.state.disabled}
						label="Name"
						help="Name of datacenter"
						type="text"
						placeholder="Enter name"
						value={datacenter.name}
						onChange={(val): void => {
							this.set('name', val);
						}}
					/>
					<PageTextArea
						label="Comment"
						help="Datacenter comment."
						placeholder="Datacenter comment"
						rows={3}
						value={datacenter.comment}
						onChange={(val: string): void => {
							this.set('comment', val);
						}}
					/>
					<PageSelect
						disabled={this.state.disabled}
						label="Network Mode"
						help="Network mode for internal VPC networking. If layer 2 networking with VLAN support isn't available VXLan must be used. A network bridge is required for the node internal interfaces when using default."
						value={datacenter.network_mode}
						onChange={(val): void => {
							this.set('network_mode', val);
						}}
					>
						<option value="default">Default</option>
						<option value="vxlan_vlan">VXLAN</option>
					</PageSelect>
					<PageSelect
						disabled={this.state.disabled}
						label="Private Storage"
						help="Private storage that will store instance snapshots."
						value={datacenter.private_storage}
						onChange={(val): void => {
							this.set('private_storage', val);
						}}
					>
						{privateStoragesSelect}
					</PageSelect>
					<PageSelect
						disabled={this.state.disabled}
						label="Private Storage Class"
						help="Private storage class to use when upload new objects."
						value={datacenter.private_storage_class}
						onChange={(val): void => {
							this.set('private_storage_class', val);
						}}
					>
						<option value="">Default</option>
						<option value="aws_standard">AWS Standard</option>
						<option value="aws_infrequent_access">AWS Standard-IA</option>
						<option value="aws_glacier">AWS Glacier</option>
					</PageSelect>
					<PageSelect
						disabled={this.state.disabled}
						label="Backup Storage"
						help="Backup storage that will store instance backups."
						value={datacenter.backup_storage}
						onChange={(val): void => {
							this.set('backup_storage', val);
						}}
					>
						{backupStoragesSelect}
					</PageSelect>
					<PageSelect
						disabled={this.state.disabled}
						label="Backup Storage Class"
						help="Backup storage class to use when upload new objects."
						value={datacenter.backup_storage_class}
						onChange={(val): void => {
							this.set('backup_storage_class', val);
						}}
					>
						<option value="">Default</option>
						<option value="aws_standard">AWS Standard</option>
						<option value="aws_infrequent_access">AWS Standard-IA</option>
						<option value="aws_glacier">AWS Glacier</option>
					</PageSelect>
				</div>
				<div style={css.group}>
					<PageInfo
						fields={[
							{
								label: 'ID',
								value: this.props.datacenter.id || 'None',
							},
						]}
					/>
					<label
						className="bp5-label"
						style={css.label}
					>
						Public Storages
						<Help
							title="Public Storages"
							content="Public storages that can be used for new instance images."
						/>
						<div>
							{publicStorages}
						</div>
					</label>
					<PageSelectButton
						label="Add Storage"
						value={this.state.addStorage}
						disabled={!hasStorages|| this.state.disabled}
						buttonClass="bp5-intent-success"
						onChange={(val: string): void => {
							this.setState({
								...this.state,
								addStorage: val,
							});
						}}
						onSubmit={this.onAddStorage}
					>
						{publicStoragesSelect}
					</PageSelectButton>
					<PageSwitch
						label="Match organizations"
						help="Limit what organizations can access this datacenter, by default all organizations will have access."
						checked={datacenter.match_organizations}
						onToggle={(): void => {
							this.toggle('match_organizations');
						}}
					/>
					<label
						className="bp5-label"
						style={css.label}
						hidden={!datacenter.match_organizations}
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
						hidden={!datacenter.match_organizations}
						buttonClass="bp5-intent-success"
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
			</div>
			<PageSave
				style={css.save}
				hidden={!this.state.datacenter}
				message={this.state.message}
				changed={this.state.changed}
				disabled={this.state.disabled}
				light={true}
				onCancel={(): void => {
					this.setState({
						...this.state,
						changed: false,
						datacenter: null,
					});
				}}
				onSave={this.onSave}
			/>
		</td>;
	}
}
