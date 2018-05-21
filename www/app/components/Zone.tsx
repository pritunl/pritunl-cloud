/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as ZoneTypes from '../types/ZoneTypes';
import * as ZoneActions from '../actions/ZoneActions';
import DatacentersStore from '../stores/DatacentersStore';
import PageInput from './PageInput';
import PageSelectButton from './PageSelectButton';
import PageInfo from './PageInfo';
import PageSave from './PageSave';
import ConfirmButton from './ConfirmButton';
import Help from './Help';

interface Props {
	zone: ZoneTypes.ZoneRo;
}

interface State {
	disabled: boolean;
	changed: boolean;
	message: string;
	zone: ZoneTypes.Zone;
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

	render(): JSX.Element {
		let zone: ZoneTypes.Zone = this.state.zone ||
			this.props.zone;

		let datacenter = DatacentersStore.datacenter(
			this.props.zone.datacenter);

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
				</div>
				<div style={css.group}>
					<PageInfo
						fields={[
							{
								label: 'ID',
								value: this.props.zone.id || 'None',
							},
							{
								label: 'Datacenter',
								value: datacenter ? datacenter.name :
									this.props.zone.datacenter || 'None',
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
