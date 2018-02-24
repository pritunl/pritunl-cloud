/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as StorageTypes from '../types/StorageTypes';
import * as StorageActions from '../actions/StorageActions';
import PageInput from './PageInput';
import PageInfo from './PageInfo';
import PageSave from './PageSave';
import PageSelect from './PageSelect';
import PageSwitch from './PageSwitch';
import ConfirmButton from './ConfirmButton';

interface Props {
	storage: StorageTypes.StorageRo;
}

interface State {
	disabled: boolean;
	changed: boolean;
	message: string;
	storage: StorageTypes.Storage;
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

export default class Storage extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			changed: false,
			message: '',
			storage: null,
		};
	}

	set(name: string, val: any): void {
		let storage: any;

		if (this.state.changed) {
			storage = {
				...this.state.storage,
			};
		} else {
			storage = {
				...this.props.storage,
			};
		}

		storage[name] = val;

		this.setState({
			...this.state,
			changed: true,
			storage: storage,
		});
	}

	onSave = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		StorageActions.commit(this.state.storage).then((): void => {
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
						storage: null,
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
		StorageActions.remove(this.props.storage.id).then((): void => {
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
		let storage: StorageTypes.Storage = this.state.storage ||
			this.props.storage;

		return <div
			className="pt-card"
			style={css.card}
		>
			<div className="layout horizontal wrap">
				<div style={css.group}>
					<div style={css.remove}>
						<ConfirmButton
							className="pt-minimal pt-intent-danger pt-icon-trash"
							progressClassName="pt-intent-danger"
							confirmMsg="Confirm storage remove"
							disabled={this.state.disabled}
							onConfirm={this.onDelete}
						/>
					</div>
					<PageInput
						disabled={this.state.disabled}
						label="Name"
						help="Name of storage"
						type="text"
						placeholder="Enter name"
						value={storage.name}
						onChange={(val): void => {
							this.set('name', val);
						}}
					/>
					<PageInput
						disabled={this.state.disabled}
						label="Endpoint"
						help="Storage endpoint domain and port"
						type="text"
						placeholder="Enter endpoint"
						value={storage.endpoint}
						onChange={(val): void => {
							this.set('endpoint', val);
						}}
					/>
					<PageInput
						disabled={this.state.disabled}
						label="Bucket"
						help="Storage bucket name"
						type="text"
						placeholder="Enter bucket"
						value={storage.bucket}
						onChange={(val): void => {
							this.set('bucket', val);
						}}
					/>
					<PageSelect
						disabled={this.state.disabled}
						label="Type"
						help="Select public for read only storages with virtual machine images. Select private for read-write storages for snapshots."
						value={storage.type}
						onChange={(val): void => {
							this.set('type', val);
						}}
					>
						<option value="public">Public</option>
						<option value="private">Private</option>
					</PageSelect>
				</div>
				<div style={css.group}>
					<PageInfo
						fields={[
							{
								label: 'ID',
								value: this.props.storage.id || 'None',
							},
						]}
					/>
					<PageInput
						disabled={this.state.disabled}
						label="Access Key"
						help="Storage access key"
						type="text"
						placeholder="Enter access key"
						value={storage.access_key}
						onChange={(val): void => {
							this.set('access_key', val);
						}}
					/>
					<PageInput
						disabled={this.state.disabled}
						label="Secret Key"
						help="Storage secret key"
						type="text"
						placeholder="Enter secret key"
						value={storage.secret_key}
						onChange={(val): void => {
							this.set('secret_key', val);
						}}
					/>
					<PageSwitch
						label="SSL Connection"
						help="Use secure SSL connection."
						disabled={this.state.disabled}
						checked={!storage.insecure}
						onToggle={(): void => {
							this.set('insecure', !storage.insecure);
						}}
					/>
				</div>
			</div>
			<PageSave
				style={css.save}
				hidden={!this.state.storage}
				message={this.state.message}
				changed={this.state.changed}
				disabled={this.state.disabled}
				light={true}
				onCancel={(): void => {
					this.setState({
						...this.state,
						changed: false,
						storage: null,
					});
				}}
				onSave={this.onSave}
			/>
		</div>;
	}
}
