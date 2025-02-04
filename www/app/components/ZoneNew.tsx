/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as ZoneTypes from '../types/ZoneTypes';
import * as ZoneActions from '../actions/ZoneActions';
import * as DatacenterTypes from '../types/DatacenterTypes';
import DatacentersStore from '../stores/DatacentersStore';
import PageInput from './PageInput';
import PageInfo from './PageInfo';
import PageCreate from './PageCreate';
import PageSelect from './PageSelect';
import ConfirmButton from './ConfirmButton';
import PageTextArea from "./PageTextArea";

interface Props {
	datacenters: DatacenterTypes.DatacentersRo;
	onClose: () => void;
}

interface State {
	closed: boolean;
	disabled: boolean;
	changed: boolean;
	message: string;
	zone: ZoneTypes.Zone;
	addCert: string;
	forwardedChecked: boolean;
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
		position: 'absolute',
		top: '5px',
		right: '5px',
	} as React.CSSProperties,
	select: {
		margin: '7px 0px 0px 6px',
		paddingTop: '3px',
	} as React.CSSProperties,
};

export default class ZoneNew extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			closed: false,
			disabled: false,
			changed: false,
			message: '',
			addCert: null,
			forwardedChecked: false,
			zone: {
				name: "New Zone",
			},
		};
	}

	set(name: string, val: any): void {
		let zone: any = {
			...this.state.zone,
		};

		zone[name] = val;

		this.setState({
			...this.state,
			changed: true,
			zone: zone,
		});
	}

	onCreate = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});

		let zone: any = {
			...this.state.zone,
		};

		ZoneActions.create(zone).then((): void => {
			this.setState({
				...this.state,
				message: 'Zone created successfully',
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
		let zone: ZoneTypes.Zone = this.state.zone;

		let hasDatacenters = false;
		let datacentersSelect: JSX.Element[] = [];
		if (this.props.datacenters && this.props.datacenters.length) {
			datacentersSelect.push(
				<option key="null" value="">Select Datacenter</option>);

			hasDatacenters = true;
			for (let datacenter of this.props.datacenters) {
				datacentersSelect.push(
					<option
						key={datacenter.id}
						value={datacenter.id}
					>{datacenter.name}</option>,
				);
			}
		}

		if (!hasDatacenters) {
			datacentersSelect.push(
				<option key="null" value="">No Datacenters</option>);
		}

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
							label="Name"
							help="Name of zone"
							type="text"
							placeholder="Enter name"
							value={zone.name}
							onChange={(val): void => {
								this.set('name', val);
							}}
						/>
						<PageTextArea
							label="Comment"
							help="Zone comment."
							placeholder="Zone comment"
							rows={3}
							value={zone.comment}
							onChange={(val: string): void => {
								this.set('comment', val);
							}}
						/>
						<PageSelect
							disabled={this.state.disabled || !hasDatacenters}
							label="Datacenter"
							help="Datacenter for zone."
							value={zone.datacenter}
							onChange={(val): void => {
								this.set("datacenter", val);
							}}
						>
							{datacentersSelect}
						</PageSelect>
						<PageSelect
							disabled={this.state.disabled}
							label="Network Mode"
							help="Network mode for internal VPC networking. If layer 2 networking with VLAN support isn't available VXLan must be used. A network bridge is required for the node internal interfaces when using default. Other zones in datacenter must use same network mode to support connectivity between zones."
							value={zone.network_mode}
							onChange={(val): void => {
								this.set('network_mode', val);
							}}
						>
							<option value="default">Default</option>
							<option value="vxlan_vlan">VXLAN</option>
						</PageSelect>
					</div>
				</div>
				<PageCreate
					style={css.save}
					hidden={!this.state.zone}
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
