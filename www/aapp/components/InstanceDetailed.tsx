/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as InstanceTypes from '../types/InstanceTypes';
import * as InstanceActions from '../actions/InstanceActions';
import PageInput from './PageInput';
import PageInfo from './PageInfo';
import PageSave from './PageSave';
import PageNumInput from './PageNumInput';
import ConfirmButton from './ConfirmButton';

interface Props {
	instance: InstanceTypes.InstanceRo;
	onClose: () => void;
}

interface State {
	disabled: boolean;
	changed: boolean;
	message: string;
	instance: InstanceTypes.Instance;
	addCert: string;
	forwardedChecked: boolean;
}

const css = {
	card: {
		position: 'relative',
		padding: '10px 10px 0 10px',
		width: '100%',
	} as React.CSSProperties,
	buttons: {
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

export default class InstanceDetailed extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			changed: false,
			message: '',
			instance: null,
			addCert: null,
			forwardedChecked: false,
		};
	}

	set(name: string, val: any): void {
		let instance: any;

		if (this.state.changed) {
			instance = {
				...this.state.instance,
			};
		} else {
			instance = {
				...this.props.instance,
			};
		}

		instance[name] = val;

		this.setState({
			...this.state,
			changed: true,
			instance: instance,
		});
	}

	onSave = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		InstanceActions.commit(this.state.instance).then((): void => {
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
						instance: null,
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
		InstanceActions.remove(this.props.instance.id).then((): void => {
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
		let instance: InstanceTypes.Instance = this.state.instance ||
			this.props.instance;

		return <td
			className="pt-cell"
			colSpan={4}
			style={css.card}
		>
			<div className="layout horizontal wrap">
				<div style={css.group}>
					<div style={css.buttons}>
						<button
							className="pt-button pt-minimal pt-intent-warning pt-icon-chevron-up"
							type="button"
							onClick={(): void => {
								this.props.onClose();
							}}
						/>
						<ConfirmButton
							className="pt-minimal pt-intent-danger pt-icon-cross"
							progressClassName="pt-intent-danger"
							confirmMsg="Confirm instance remove"
							disabled={this.state.disabled}
							onConfirm={this.onDelete}
						/>
					</div>
					<PageInput
						label="Name"
						help="Name of instance"
						type="text"
						placeholder="Enter name"
						value={instance.name}
						onChange={(val): void => {
							this.set('name', val);
						}}
					/>
					<PageNumInput
						label="Memory Size"
						help="Instance memory size in megabytes."
						min={256}
						minorStepSize={256}
						stepSize={512}
						majorStepSize={1024}
						disabled={this.state.disabled}
						selectAllOnFocus={true}
						onChange={(val: number): void => {
							this.set('memory', val);
						}}
						value={instance.memory}
					/>
					<PageNumInput
						label="Processors"
						help="Number of instance processors."
						min={1}
						minorStepSize={1}
						stepSize={1}
						majorStepSize={2}
						disabled={this.state.disabled}
						selectAllOnFocus={true}
						onChange={(val: number): void => {
							this.set('processors', val);
						}}
						value={instance.processors}
					/>
				</div>
				<div style={css.group}>
					<PageInfo
						fields={[
							{
								label: 'ID',
								value: instance.id || 'None',
							},
							{
								label: 'Zone',
								value: instance.zone || 'None',
							},
						]}
					/>
				</div>
			</div>
			<PageSave
				style={css.save}
				hidden={!this.state.instance}
				message={this.state.message}
				changed={this.state.changed}
				disabled={this.state.disabled}
				light={true}
				onCancel={(): void => {
					this.setState({
						...this.state,
						changed: false,
						forwardedChecked: false,
						instance: null,
					});
				}}
				onSave={this.onSave}
			/>
		</td>;
	}
}
