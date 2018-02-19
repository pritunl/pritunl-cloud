/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as DatacenterTypes from '../types/DatacenterTypes';
import * as StorageTypes from '../types/StorageTypes';
import * as DatacenterActions from '../actions/DatacenterActions';
import StoragesStore from '../stores/StoragesStore';
import PageInput from './PageInput';
import PageInfo from './PageInfo';
import PageSelectButton from './PageSelectButton';
import PageSave from './PageSave';
import ConfirmButton from './ConfirmButton';
import Help from './Help';

interface Props {
	datacenter: DatacenterTypes.DatacenterRo;
	storages: StorageTypes.StoragesRo;
}

interface State {
	disabled: boolean;
	changed: boolean;
	message: string;
	datacenter: DatacenterTypes.Datacenter;
	addStorage: string;
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

export default class Datacenter extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			changed: false,
			message: '',
			datacenter: null,
			addStorage: '',
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

		let storageId = this.state.addStorage ||
			this.props.storages[0].id;

		if (this.state.changed) {
			datacenter = {
				...this.state.datacenter,
			};
		} else {
			datacenter = {
				...this.props.datacenter,
			};
		}

		let storages = [
			...(datacenter.storages || []),
		];

		if (storages.indexOf(storageId) === -1) {
			storages.push(storageId);
		}

		storages.sort();

		datacenter.storages = storages;

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

		let storages = [
			...(datacenter.storages || []),
		];

		let i = storages.indexOf(storage);
		if (i === -1) {
			return;
		}

		storages.splice(i, 1);

		datacenter.storages = storages;

		this.setState({
			...this.state,
			changed: true,
			datacenter: datacenter,
		});
	}

	render(): JSX.Element {
		let datacenter: DatacenterTypes.Datacenter = this.state.datacenter ||
			this.props.datacenter;

		let storages: JSX.Element[] = [];
		for (let storageId of (datacenter.storages || [])) {
			let storage = StoragesStore.storage(storageId);
			if (!storage) {
				continue;
			}

			storages.push(
				<div
					className="pt-tag pt-tag-removable pt-intent-primary"
					style={css.item}
					key={storage.id}
				>
					{storage.name}
					<button
						className="pt-tag-remove"
						onMouseUp={(): void => {
							this.onRemoveStorage(storage.id);
						}}
					/>
				</div>,
			);
		}

		let storagesSelect: JSX.Element[] = [];
		if (this.props.storages.length) {
			for (let storage of this.props.storages) {
				storagesSelect.push(
					<option
						key={storage.id}
						value={storage.id}
					>{storage.name}</option>,
				);
			}
		} else {
			storagesSelect.push(<option key="null" value="">None</option>);
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
							confirmMsg="Confirm datacenter remove"
							disabled={this.state.disabled}
							onConfirm={this.onDelete}
						/>
					</div>
					<PageInput
						label="Name"
						help="Name of datacenter"
						type="text"
						placeholder="Enter name"
						value={datacenter.name}
						onChange={(val): void => {
							this.set('name', val);
						}}
					/>
					<label
						className="pt-label"
						style={css.label}
					>
						Storages
						<Help
							title="Storages"
							content="Storages that can access this zone."
						/>
						<div>
							{storages}
						</div>
					</label>
					<PageSelectButton
						label="Add Storage"
						value={this.state.addStorage}
						disabled={!this.props.storages.length}
						buttonClass="pt-intent-success"
						onChange={(val: string): void => {
							this.setState({
								...this.state,
								addStorage: val,
							});
						}}
						onSubmit={this.onAddStorage}
					>
						{storagesSelect}
					</PageSelectButton>
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
		</div>;
	}
}
