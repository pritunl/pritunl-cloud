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
import PageSelect from "./PageSelect";
import PageTextArea from "./PageTextArea";

interface Props {
	block: BlockTypes.BlockRo;
	selected: boolean;
	onSelect: (shift: boolean) => void;
	onClose: () => void;
}

interface State {
	disabled: boolean;
	changed: boolean;
	message: string;
	addSubnet: string,
	addSubnet6: string,
	addExclude: string,
	block: BlockTypes.Block;
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
	controlButton: {
		marginRight: '10px',
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
		backgroundColor: 'rgba(0, 0, 0, 0.13)',
	} as React.CSSProperties,
	select: {
		margin: '7px 0px 0px 6px',
		paddingTop: '3px',
	} as React.CSSProperties,
};

export default class BlockDetailed extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			changed: false,
			message: '',
			addSubnet: '',
			addSubnet6: '',
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
			block: block,
		});
	}

	onAddSubnet6 = (): void => {
		let block: BlockTypes.Block;

		if (!this.state.addSubnet6) {
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

		let subnets6 = [
			...(block.subnets6 || []),
		];

		let addSubnet6 = this.state.addSubnet6.trim();
		if (subnets6.indexOf(addSubnet6) === -1) {
			subnets6.push(addSubnet6);
		}

		subnets6.sort();
		block.subnets6 = subnets6;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addSubnet6: '',
			block: block,
		});
	}

	onRemoveSubnet6 = (subnet: string): void => {
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

		let subnets6 = [
			...(block.subnets6 || []),
		];

		let i = subnets6.indexOf(subnet);
		if (i === -1) {
			return;
		}

		subnets6.splice(i, 1);
		block.subnets6 = subnets6;

		this.setState({
			...this.state,
			changed: true,
			message: '',
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
					className="bp5-tag bp5-tag-removable bp5-intent-primary"
					style={css.item}
					key={subnet}
				>
					{subnet}
					<button
						className="bp5-tag-remove"
						disabled={this.state.disabled}
						onMouseUp={(): void => {
							this.onRemoveSubnet(subnet);
						}}
					/>
				</div>,
			);
		}

		let subnets6: JSX.Element[] = [];
		for (let subnet of (block.subnets6 || [])) {
			subnets6.push(
				<div
					className="bp5-tag bp5-tag-removable bp5-intent-primary"
					style={css.item}
					key={subnet}
				>
					{subnet}
					<button
						className="bp5-tag-remove"
						disabled={this.state.disabled}
						onMouseUp={(): void => {
							this.onRemoveSubnet6(subnet);
						}}
					/>
				</div>,
			);
		}

		let excludes: JSX.Element[] = [];
		for (let exclude of (block.excludes || [])) {
			excludes.push(
				<div
					className="bp5-tag bp5-tag-removable bp5-intent-primary"
					style={css.item}
					key={exclude}
				>
					{exclude}
					<button
						className="bp5-tag-remove"
						disabled={this.state.disabled}
						onMouseUp={(): void => {
							this.onRemoveExclude(exclude);
						}}
					/>
				</div>,
			);
		}

		return <td
			className="bp5-cell"
			colSpan={2}
			style={css.card}
		>
			<div className="layout horizontal wrap">
				<div style={css.group}>
					<div
						className="layout horizontal tab-close"
						style={css.buttons}
						onClick={(evt): void => {
							let target = evt.target as HTMLElement;

							if (target.className.indexOf('tab-close') !== -1) {
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
							dialogLabel="Delete Block"
							confirmMsg="Permanently delete this block"
							confirmInput={true}
							items={[block.name]}
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
					<PageTextArea
						label="Comment"
						help="Block comment."
						placeholder="Block comment"
						rows={3}
						value={block.comment}
						onChange={(val: string): void => {
							this.set('comment', val);
						}}
					/>
					<PageSelect
						disabled={true}
						label="Network Mode"
						help="Network mode IP block."
						value={block.type}
						onChange={(val): void => {
							this.set('type', val);
						}}
					>
						<option value="ipv4">IPv4</option>
						<option value="ipv6">IPv6</option>
					</PageSelect>
					<PageInput
						disabled={this.state.disabled}
						hidden={block.type === 'ipv6'}
						label="Netmask"
						help="Netmask of IP block"
						type="text"
						placeholder="Enter netmask"
						value={block.netmask}
						onChange={(val): void => {
							this.set('netmask', val);
						}}
					/>
					<label
						className="bp5-label"
						hidden={block.type === 'ipv6'}
					>
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
						hidden={block.type === 'ipv6'}
						buttonClass="bp5-intent-success bp5-icon-add"
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
					<label
						className="bp5-label"
						hidden={block.type !== 'ipv6'}
					>
						IPv6 Addresses
						<Help
							title="IPv6 Addresses"
							content="IPv6 addresses that are available for instances."
						/>
						<div>
							{subnets6}
						</div>
					</label>
					<PageInputButton
						disabled={this.state.disabled}
						hidden={block.type !== 'ipv6'}
						buttonClass="bp5-intent-success bp5-icon-add"
						label="Add"
						type="text"
						placeholder="Add addresses"
						value={this.state.addSubnet6}
						onChange={(val): void => {
							this.setState({
								...this.state,
								addSubnet6: val,
							});
						}}
						onSubmit={this.onAddSubnet6}
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
						hidden={block.type === 'ipv6'}
						label="Gateway"
						help="Gateway address of IP block"
						type="text"
						placeholder="Enter gateway"
						value={block.gateway}
						onChange={(val): void => {
							this.set('gateway', val);
						}}
					/>
					<PageInput
						disabled={this.state.disabled}
						hidden={block.type !== 'ipv6'}
						label="IPv6 Gateway"
						help="Gateway address of IPv6 block"
						type="text"
						placeholder="Enter IPv6 gateway"
						value={block.gateway6}
						onChange={(val): void => {
							this.set('gateway6', val);
						}}
					/>
					<label
						className="bp5-label"
						hidden={block.type === 'ipv6'}
					>
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
						hidden={block.type === 'ipv6'}
						buttonClass="bp5-intent-success bp5-icon-add"
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
		</td>;
	}
}
