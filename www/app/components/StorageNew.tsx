/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as StorageTypes from '../types/StorageTypes';
import * as StorageActions from '../actions/StorageActions';
import PageInput from './PageInput';
import PageInfo from './PageInfo';
import PageCreate from './PageCreate';
import PageSelect from './PageSelect';
import PageSwitch from './PageSwitch';
import ConfirmButton from './ConfirmButton';
import * as Alert from "../Alert";
import PageTextArea from "./PageTextArea";

interface Props {
	onClose: () => void;
}

interface State {
	closed: boolean;
	disabled: boolean;
	changed: boolean;
	message: string;
	storage: StorageTypes.Storage;
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
	item: {
		margin: '9px 5px 0 5px',
		minHeight: '20px',
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
	controlButton: {
		marginRight: '10px',
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

export default class StorageDetailed extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			closed: false,
			disabled: false,
			changed: false,
			message: '',
			storage: {
				name: "new-storage",
			},
		};
	}

	set(name: string, val: any): void {
		let storage: any = {
			...this.state.storage,
		};

		storage[name] = val;

		this.setState({
			...this.state,
			changed: true,
			storage: storage,
		});
	}

	onCreate = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});

		let storage: any = {
			...this.state.storage,
		};

		StorageActions.create(storage).then((): void => {
			this.setState({
				...this.state,
				message: 'Storage created successfully',
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
		let storage: StorageTypes.Storage = this.state.storage;

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
						<PageTextArea
							label="Comment"
							help="Storage comment."
							placeholder="Storage comment"
							rows={3}
							value={storage.comment}
							onChange={(val: string): void => {
								this.set('comment', val);
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
					</div>
					<div style={css.group}>
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
							<option value="web">Web</option>
						</PageSelect>
						<PageInput
							disabled={this.state.disabled}
							hidden={storage.type == "web"}
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
							hidden={storage.type == "web"}
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
				<PageCreate
					style={css.save}
					hidden={!this.state.storage}
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
