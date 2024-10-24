/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Blueprint from '@blueprintjs/core';
import * as Icons from '@blueprintjs/icons';
import * as Constants from '../Constants';
import * as ServiceTypes from '../types/ServiceTypes';
import * as ServiceActions from '../actions/ServiceActions';
import ServicesUnitStore from '../stores/ServicesUnitStore';
import * as MiscUtils from '../utils/MiscUtils';
import * as Theme from '../Theme';
import PageInput from './PageInput';
import PageSelect from './PageSelect';
import PageInfo from './PageInfo';
import PageInputButton from './PageInputButton';
import ServiceEditor from './ServiceEditor';
import ServiceUnit from './ServiceUnit';
import PageSave from './PageSave';
import ConfirmButton from './ConfirmButton';
import Help from './Help';
import PageTextArea from "./PageTextArea";

interface Props {
	service: ServiceTypes.ServiceRo;
	disabled: boolean;
	unitChanged: boolean;
	mode: string;
	onMode: (mode: string) => void;
	onChange: (units: ServiceTypes.Unit[]) => void;
}

interface State {
	expandLeft: boolean;
	expandRight: boolean;
	activeUnitId: string;
	unit: ServiceTypes.ServiceUnit;
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
		padding: "0 0 0 1px",
	} as React.CSSProperties,
	tab: {
		fontWeight: "bold",
		marginRight: "10px",
	} as React.CSSProperties,
	documentIcon: {
		margin: "2px 0 0 0",
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
	divider : {
		marginRight: "0",
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
	navButton: {
		marginLeft: '10px',
	} as React.CSSProperties,
	settingsOpen: {
		marginLeft: '10px',
	} as React.CSSProperties,
	settingsMenu: {
		maxHeight: '400px',
		overflowY: "scroll",
	} as React.CSSProperties,
};

export default class ServiceWorkspace extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			expandLeft: true,
			expandRight: false,
			activeUnitId: "",
			unit: null,
		};
	}

	componentDidMount(): void {
		ServicesUnitStore.addChangeListener(this.onChange);
		let activeUnit = this.getActiveUnit()
		if (activeUnit && !activeUnit.new) {
			ServiceActions.syncUnit(this.props.service.id, activeUnit.id);
		}
	}

	componentWillUnmount(): void {
		ServicesUnitStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		let unit: ServiceTypes.ServiceUnit
		let activeUnit = this.getActiveUnit()

		if (activeUnit && !activeUnit.new) {
			unit = ServicesUnitStore.unit(activeUnit.id)
		} else {
			unit = null
		}

		this.setState({
			...this.state,
			unit: unit,
		});
	}

	getActiveUnit = (): ServiceTypes.Unit => {
		let units = [
			...(this.props.service.units || []),
		]

		let activeUnit = units.find(unit => unit.id === this.state.activeUnitId)
		if (!activeUnit) {
			for (let unit of units) {
				if (!unit.delete) {
					activeUnit = unit
					break
				}
			}
		}

		return activeUnit
	}

	getActiveUnitIndex = (): number => {
		let units = [
			...(this.props.service.units || []),
		]

		let activeIndex = units.findIndex(
			unit => unit.id === this.state.activeUnitId)
		if (activeIndex === -1) {
			for (let i = 0; i < units.length; i++) {
				if (!units[i].delete) {
					activeIndex = i
					break
				}
			}
		}

		return activeIndex
	}

	onEdit = (): void => {
		let units = [
			...(this.props.service.units || []),
		]

		this.setState({
			...this.state,
			expandLeft: false,
			expandRight: true,
		})
		this.props.onChange(units)
	}

	onView = (): void => {
		this.setState({
			...this.state,
			expandLeft: true,
			expandRight: false,
		})
		this.props.onMode("view")
	}

	onUnit = (): void => {
		this.setState({
			...this.state,
			expandLeft: true,
			expandRight: false,
		})
		this.props.onMode("unit")
	}

	onNew = (): void => {
		let units = [
			...(this.props.service.units || []),
		]

		units.push({
			id: MiscUtils.objectId(),
			name: "new-unit",
			spec: "",
			new: true,
		})

		this.setState({
			...this.state,
			expandLeft: false,
			expandRight: true,
		})
		this.props.onChange(units)
	}

	onDelete = (): void => {
		let units = [
			...(this.props.service.units || []),
		]

		let index = this.getActiveUnitIndex()
		if (index !== -1) {
			let unit = units[index]
			units[index] = {
				id: unit.id,
				delete: true,
			}
		}

		this.setState({
			...this.state,
			activeUnitId: "",
		})
		this.props.onChange(units)

		let activeUnit = this.getActiveUnit()
		if (activeUnit && !activeUnit.new) {
			ServiceActions.syncUnit(this.props.service.id, activeUnit.id);
		}
	}

	onUnitEdit = (val: string): void => {
		let units = [
			...(this.props.service.units || []),
		]

		let index = this.getActiveUnitIndex()
		if (index !== -1) {
			units[index].spec = val
		}

		this.props.onChange(units)
	}

	render(): JSX.Element {
		let units = [
			...(this.props.service.units || []),
		]
		let activeUnit = this.getActiveUnit()

		let expandLeft = this.state.expandLeft
		let expandRight = this.state.expandRight
		if (!this.props.unitChanged) {
			expandLeft = true
			expandRight = false
		}

		let expandIconClass: string
		if (!expandLeft && !expandRight) {
			expandIconClass = "bp5-button bp5-minimal bp5-icon-maximize"
		} else {
			expandIconClass = "bp5-button bp5-minimal bp5-icon-minimize"
		}

		let tabsElem: JSX.Element[] = []
		for (let i = 0; i < units.length; ++i) {
			let unit = units[i]
			if (unit.delete) {
				continue
			}

			tabsElem.push(<Blueprint.Tab id={unit.id} style={css.tab} key={unit.id}>
				{unit.name}
			</Blueprint.Tab>)
		}

		let curEditorTheme = Theme.getEditorTheme()
		let fontMenuItems: JSX.Element[] = []
		for (let editorTheme in Theme.editorThemeNames) {
			let className = ""
			let themeName = Theme.editorThemeNames[editorTheme]

			if (editorTheme === curEditorTheme) {
				className = "bp5-intent-primary"
			}

			let menuItem = <Blueprint.MenuItem
				key={editorTheme}
				className={className}
				icon={<Icons.Font/>}
				onClick={(): void => {
					Theme.setEditorTheme(editorTheme)
					Theme.save()
					this.setState({
						...this.state,
					})
				}}
				text={themeName}
			/>
			fontMenuItems.push(menuItem)
		}

		let settingsMenu = <Blueprint.Menu style={css.settingsMenu}>
			<li className="bp5-menu-header">
				<h6 className="bp5-heading">Editor Theme</h6>
			</li>
			<Blueprint.MenuDivider/>
			{fontMenuItems}
		</Blueprint.Menu>

		return <div
			style={css.card}
		>
			<Blueprint.Navbar>
				<Blueprint.NavbarGroup align={"left"}>
					<Blueprint.Tabs
						id={this.props.service.id}
						selectedTabId={activeUnit ? activeUnit.id : null}
						fill={true}
						onChange={(newTabId): void => {
							let activeUnitId = newTabId.valueOf() as string
							let activeUnit = units.find(unit => unit.id === activeUnitId)

							this.setState({
								...this.state,
								activeUnitId: activeUnitId,
							})

							if (activeUnit && !activeUnit.new) {
								ServiceActions.syncUnit(this.props.service.id, activeUnitId);
							}
						}}
					>
						{tabsElem}
					</Blueprint.Tabs>
				</Blueprint.NavbarGroup>
				<Blueprint.NavbarGroup align={"right"}>
					<Blueprint.NavbarDivider
						style={css.divider}
					/>
					<button
						hidden={this.props.mode !== "edit"}
						style={css.navButton}
						className={expandIconClass}
						onClick={(): void => {
							this.setState({
								...this.state,
								expandLeft: false,
								expandRight: !expandRight,
							})
						}}
					/>
					<Blueprint.Popover
						content={settingsMenu}
						placement="bottom"
					>
						<Blueprint.Button
							alignText="left"
							icon={<Icons.Application/>}
							rightIcon={<Icons.CaretDown/>}
							text="Settings"
							style={css.settingsOpen}
							hidden={this.props.mode !== "edit"}
						/>
					</Blueprint.Popover>
					<button
						disabled={this.props.disabled}
						hidden={this.props.mode === "view"}
						style={css.navButton}
						className="bp5-button bp5-icon-document-open"
						onClick={(): void => {
							this.onView()
						}}
					>View Spec</button>
					<button
						disabled={this.props.disabled}
						hidden={this.props.mode === "edit" || !activeUnit}
						style={css.navButton}
						className="bp5-button bp5-icon-edit"
						onClick={(): void => {
							this.onEdit()
						}}
					>Edit Spec</button>
					<button
						disabled={this.props.disabled}
						hidden={this.props.mode === "unit"}
						style={css.navButton}
						className="bp5-button bp5-icon-dashboard"
						onClick={(): void => {
							this.onUnit()
						}}
					>Deployments</button>
					<button
						disabled={this.props.disabled}
						style={css.navButton}
						className="bp5-button bp5-icon-plus"
						onClick={(): void => {
							this.onNew()
						}}
					>New Unit</button>
					<ConfirmButton
						safe={true}
						className="bp5-intent-danger bp5-icon-trash"
						progressClassName="bp5-intent-danger"
						dialogClassName="bp5-intent-danger bp5-icon-delete"
						dialogLabel="Delete Unit"
						label="Delete Unit"
						confirmMsg="Permanently delete this unit"
						confirmInput={false}
						items={[activeUnit ? activeUnit.name : "null"]}
						hidden={!activeUnit}
						style={css.navButton}
						disabled={this.props.disabled}
						onConfirm={(): void => {
							this.onDelete()
						}}
					/>
				</Blueprint.NavbarGroup>
			</Blueprint.Navbar>
			<ServiceEditor
				hidden={this.props.mode === "unit"}
				expandLeft={expandLeft}
				expandRight={expandRight}
				disabled={this.props.disabled}
				readOnly={this.props.mode === "view"}
				uuid={activeUnit ? activeUnit.id : null}
				value={activeUnit ? activeUnit.spec : null}
				onChange={(val: string): void => {
					this.onUnitEdit(val)
				}}
				onEdit={this.onEdit}
			/>
			<ServiceUnit
				hidden={this.props.mode !== "unit"}
				disabled={this.props.disabled}
				unit={this.state.unit}
			/>
		</div>;
	}
}
