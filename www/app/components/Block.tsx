/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as BlockTypes from '../types/BlockTypes';
import * as BlockActions from '../actions/BlockActions';
import PageInput from './PageInput';
import PageInfo from './PageInfo';
import PageSave from './PageSave';
import PageInputButton from './PageInputButton';
import ConfirmButton from './ConfirmButton';
import * as Alert from "../Alert";
import Help from './Help';

interface Props {
	block: BlockTypes.BlockRo;
}

interface State {
	disabled: boolean;
	changed: boolean;
	message: string;
	addSubnet: string,
	addExclude: string,
	block: BlockTypes.Block;
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
};

export default class Block extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			changed: false,
			message: '',
			addSubnet: '',
			addExclude: '',
			block: null,
		};
	}

	set(name: string, val: any): void {
		let block: any;

		if (this.state.changed) {
			block = {
				...this.state.block,
			};
		} else {
			block = {
				...this.props.block,
			};
		}

		block[name] = val;

		this.setState({
			...this.state,
			changed: true,
			block: block,
		});
	}

	onSave = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		BlockActions.commit(this.state.block).then((): void => {
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
						block: null,
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

	onSync = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		BlockActions.commit(this.props.block).then((): void => {
			this.setState({
				...this.state,
				disabled: false,
			});

			Alert.success('Block sync started');
		}).catch((): void => {
			this.setState({
				...this.state,
				disabled: false,
			});
		});
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		BlockActions.remove(this.props.block.id).then((): void => {
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

	onAddSubnet = (): void => {
		let block: BlockTypes.Block;

		if (!this.state.addSubnet) {
			return;
		}

		if (this.state.changed) {
			block = {
				...this.state.block,
			};
		} else {
			block = {
				...this.props.block,
			};
		}

		let subnets = [
			...(block.subnets || []),
		];

		let addSubnet = this.state.addSubnet.trim();
		if (subnets.indexOf(addSubnet) === -1) {
			subnets.push(addSubnet);
		}

		subnets.sort();
		block.subnets = subnets;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addSubnet: '',
			block: block,
		});
	}

	onRemoveSubnet = (subnet: string): void => {
		let block: BlockTypes.Block;

		if (this.state.changed) {
			block = {
				...this.state.block,
			};
		} else {
			block = {
				...this.props.block,
			};
		}

		let subnets = [
			...(block.subnets || []),
		];

		let i = subnets.indexOf(subnet);
		if (i === -1) {
			return;
		}

		subnets.splice(i, 1);
		block.subnets = subnets;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addSubnet: '',
			block: block,
		});
	}

	onAddExclude = (): void => {
		let block: BlockTypes.Block;

		if (!this.state.addExclude) {
			return;
		}

		if (this.state.changed) {
			block = {
				...this.state.block,
			};
		} else {
			block = {
				...this.props.block,
			};
		}

		let excludes = [
			...(block.excludes || []),
		];

		let addExclude = this.state.addExclude.trim();
		if (excludes.indexOf(addExclude) === -1) {
			excludes.push(addExclude);
		}

		excludes.sort();
		block.excludes = excludes;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addExclude: '',
			block: block,
		});
	}

	onRemoveExclude = (exclude: string): void => {
		let block: BlockTypes.Block;

		if (this.state.changed) {
			block = {
				...this.state.block,
			};
		} else {
			block = {
				...this.props.block,
			};
		}

		let excludes = [
			...(block.excludes || []),
		];

		let i = excludes.indexOf(exclude);
		if (i === -1) {
			return;
		}

		excludes.splice(i, 1);
		block.excludes = excludes;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addExclude: '',
			block: block,
		});
	}

	render(): JSX.Element {
		let block: BlockTypes.Block = this.state.block ||
			this.props.block;

		let subnets: JSX.Element[] = [];
		for (let subnet of (block.subnets || [])) {
			subnets.push(
				<div
					className="bp3-tag bp3-tag-removable bp3-intent-primary"
					style={css.item}
					key={subnet}
				>
					{subnet}
					<button
						className="bp3-tag-remove"
						disabled={this.state.disabled}
						onMouseUp={(): void => {
							this.onRemoveSubnet(subnet);
						}}
					/>
				</div>,
			);
		}

		let excludes: JSX.Element[] = [];
		for (let exclude of (block.excludes || [])) {
			excludes.push(
				<div
					className="bp3-tag bp3-tag-removable bp3-intent-primary"
					style={css.item}
					key={exclude}
				>
					{exclude}
					<button
						className="bp3-tag-remove"
						disabled={this.state.disabled}
						onMouseUp={(): void => {
							this.onRemoveExclude(exclude);
						}}
					/>
				</div>,
			);
		}

		return <div
			className="bp3-card"
			style={css.card}
		>
			<div className="layout horizontal wrap">
				<div style={css.group}>
					<div style={css.remove}>
						<ConfirmButton
							className="bp3-minimal bp3-intent-danger bp3-icon-trash"
							progressClassName="bp3-intent-danger"
							confirmMsg="Confirm block remove"
							disabled={this.state.disabled}
							onConfirm={this.onDelete}
						/>
					</div>
					<PageInput
						disabled={this.state.disabled}
						label="Name"
						help="Name of IP block"
						type="text"
						placeholder="Enter name"
						value={block.name}
						onChange={(val): void => {
							this.set('name', val);
						}}
					/>
					<label className="bp3-label">
						IP Addresses
						<Help
							title="IP Addresses"
							content="IP addresses that are available for instances."
						/>
						<div>
							{subnets}
						</div>
					</label>
					<PageInputButton
						disabled={this.state.disabled}
						buttonClass="bp3-intent-success bp3-icon-add"
						label="Add"
						type="text"
						placeholder="Add addresses"
						value={this.state.addSubnet}
						onChange={(val): void => {
							this.setState({
								...this.state,
								addSubnet: val,
							});
						}}
						onSubmit={this.onAddSubnet}
					/>
					<label className="bp3-label">
						IP Excludes
						<Help
							title="IP Excludes"
							content="IP addresses that are excluded from block. Add host or other reserved addresses here."
						/>
						<div>
							{excludes}
						</div>
					</label>
					<PageInputButton
						disabled={this.state.disabled}
						buttonClass="bp3-intent-success bp3-icon-add"
						label="Add"
						type="text"
						placeholder="Add exclude"
						value={this.state.addExclude}
						onChange={(val): void => {
							this.setState({
								...this.state,
								addExclude: val,
							});
						}}
						onSubmit={this.onAddExclude}
					/>
				</div>
				<div style={css.group}>
					<PageInfo
						fields={[
							{
								label: 'ID',
								value: this.props.block.id || 'None',
							},
						]}
					/>
					<PageInput
						disabled={this.state.disabled}
						label="Netmask"
						help="Netmask of IP block"
						type="text"
						placeholder="Enter netmask"
						value={block.netmask}
						onChange={(val): void => {
							this.set('netmask', val);
						}}
					/>
					<PageInput
						disabled={this.state.disabled}
						label="Gateway"
						help="Gateway address of IP block"
						type="text"
						placeholder="Enter gateway"
						value={block.gateway}
						onChange={(val): void => {
							this.set('gateway', val);
						}}
					/>
				</div>
			</div>
			<PageSave
				style={css.save}
				hidden={!this.state.block}
				message={this.state.message}
				changed={this.state.changed}
				disabled={this.state.disabled}
				light={true}
				onCancel={(): void => {
					this.setState({
						...this.state,
						changed: false,
						block: null,
					});
				}}
				onSave={this.onSave}
			/>
		</div>;
	}
}
