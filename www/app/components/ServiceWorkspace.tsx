/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Blueprint from '@blueprintjs/core';
import * as Constants from '../Constants';
import * as ServiceTypes from '../types/ServiceTypes';
import * as ServiceActions from '../actions/ServiceActions';
import * as OrganizationTypes from "../types/OrganizationTypes";
import * as MiscUtils from '../utils/MiscUtils';
import PageInput from './PageInput';
import PageSelect from './PageSelect';
import PageInfo from './PageInfo';
import PageInputButton from './PageInputButton';
import ServiceEditor from './ServiceEditor';
import PageSave from './PageSave';
import ConfirmButton from './ConfirmButton';
import Help from './Help';
import PageTextArea from "./PageTextArea";
import * as DomainTypes from "../types/DomainTypes";
import * as VpcTypes from "../types/VpcTypes";
import * as DatacenterTypes from "../types/DatacenterTypes";
import * as NodeTypes from "../types/NodeTypes";
import * as PoolTypes from "../types/PoolTypes";
import * as ZoneTypes from "../types/ZoneTypes";
import * as ShapeTypes from "../types/ShapeTypes";
import * as Theme from "../Theme";

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	domains: DomainTypes.DomainsRo;
	vpcs: VpcTypes.VpcsRo;
	datacenters: DatacenterTypes.DatacentersRo;
	nodes: NodeTypes.NodesRo;
	pools: PoolTypes.PoolsRo;
	zones: ZoneTypes.ZonesRo;
	shapes: ShapeTypes.ShapesRo;
	service: ServiceTypes.ServiceRo;
	disabled: boolean;
	onEdit?: () => void;
}

interface State {
	units: ServiceTypes.Unit[]
	readOnly: boolean;
	expandLeft: boolean;
	expandRight: boolean;
	activeUnit: number;
}

const css = {
	card: {
		position: 'relative',
		padding: '48px 10px 0 10px',
		width: '100%',
	} as React.CSSProperties,
	editButton: {
		margin: "0 0 0 3px",
		minHeight: "18px",
		minWidth: "18px",
		height: "18px",
		width: "18px",
		padding: "4px 0 0 4px",
	} as React.CSSProperties,
	tab: {
		marginRight: "10px",
	} as React.CSSProperties,
	documentIcon: {
		margin: "2px 0 0 0",
		fontSize: "12px",
	} as React.CSSProperties,
	editButtonIcon: {
		fontSize: "12px",
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
	status: {
		margin: '6px 0 0 1px',
	} as React.CSSProperties,
	icon: {
		marginRight: '3px',
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
	select: {
		margin: '7px 0px 0px 6px',
		paddingTop: '3px',
	} as React.CSSProperties,
	role: {
		margin: '9px 5px 0 5px',
		height: '20px',
	} as React.CSSProperties,
	rules: {
		marginBottom: '15px',
	} as React.CSSProperties,
};

export default class ServiceWorkspace extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			units: null,
			readOnly: true,
			expandLeft: true,
			expandRight: false,
			activeUnit: 0,
		};
	}

	onEdit = (): void => {
		let units = this.props.service.units || []

		for (let unit of units) {
			if (!unit.id) {
				unit.id = MiscUtils.uuid()
			}
		}

		this.setState({
			...this.state,
			readOnly: false,
			expandLeft: false,
			expandRight: true,
			units: units,
		})
	}

	onChange = (val: string): void => {
		let units = this.props.service.units || []

		for (let unit of units) {
			if (!unit.id) {
				unit.id = MiscUtils.uuid()
			}
		}

		units[this.state.activeUnit].spec = val

		this.setState({
			...this.state,
			units: units,
		})
	}

	render(): JSX.Element {
		let units = this.state.units || this.props.service.units || [];

		let expandLeft = this.state.expandLeft
		let expandRight = this.state.expandRight
		let expandIconClass: string

		if (!expandLeft && !expandRight) {
			expandIconClass = "bp5-button bp5-minimal bp5-icon-maximize"
		} else {
			expandIconClass = "bp5-button bp5-minimal bp5-icon-minimize"
		}

		let tabsElem: JSX.Element[] = []
		for (let i = 0; i < units.length; ++i) {
			let unit = units[i]
			tabsElem.push(<Blueprint.Tab id={i} style={css.tab}>
				<Blueprint.Icon icon="document" style={css.editButtonIcon}/>
				{unit.name}
				<button
					className="bp5-button bp5-minimal bp5-small"
					type="button"
					style={css.editButton}
					disabled={this.props.disabled}
					onClick={(): void => {
						console.log("test")
					}}
				><Blueprint.Icon icon="cog" style={css.editButtonIcon}/></button>
			</Blueprint.Tab>)
		}

		return <div
			style={css.card}
		>
			<Blueprint.Navbar>
				<Blueprint.NavbarGroup align={"left"}>
					<Blueprint.Tabs
						id="TODO"
						fill={true}
						onChange={(newTabId, prevTabId): void => {
							this.setState({
								...this.state,
								activeUnit: newTabId.valueOf() as number,
							})
						}}
					>
						{tabsElem}
					</Blueprint.Tabs>
				</Blueprint.NavbarGroup>
				<Blueprint.NavbarGroup align={"right"}>
					<Blueprint.NavbarDivider/>
					<button
						disabled={this.props.disabled}
						hidden={!this.state.readOnly}
						className="bp5-button bp5-icon-edit"
						onClick={(): void => {
							this.onEdit()
						}}
					>Edit Spec</button>
					<button
						hidden={this.state.readOnly}
						className={expandIconClass}
						onClick={(): void => {
							this.setState({
								...this.state,
								expandLeft: false,
								expandRight: !expandRight,
							})
						}}
					/>
				</Blueprint.NavbarGroup>
			</Blueprint.Navbar>
			<ServiceEditor
				expandLeft={expandLeft}
				expandRight={expandRight}
				disabled={this.props.disabled}
				readOnly={this.state.readOnly}
				uuid={units[this.state.activeUnit].id}
				value={units[this.state.activeUnit].spec}
				onChange={(val: string): void => {
					this.onChange(val)
				}}
				onEdit={this.onEdit}
			/>
		</div>;
	}
}
