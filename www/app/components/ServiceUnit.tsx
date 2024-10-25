/// <reference path="../References.d.ts"/>
import * as React from "react"
import * as Blueprint from "@blueprintjs/core"
import Help from "./Help"
import * as ServiceTypes from "../types/ServiceTypes"
import * as ServiceActions from "../actions/ServiceActions"
import PageInfo from "./PageInfo"

interface Props {
	hidden: boolean
	disabled?: boolean
	unit: ServiceTypes.ServiceUnit
}

interface State {
}

const css = {
	box: {
		flex: 1,
		minWidth: "280px",
		margin: "10px 10px 20px 10px",
		fontSize: "12px",
	} as React.CSSProperties,
	card: {
		padding: "5px 5px 3px 5px",
	} as React.CSSProperties,
	check: {
		margin: "0 0 0 0",
	} as React.CSSProperties,
	checkBox: {
		display: "flex",
		paddingBottom: "2px",
	} as React.CSSProperties,
	info: {
		marginBottom: "0px",
	} as React.CSSProperties,
	item1: {
		flex: 1,
		minWidth: "150px",
	} as React.CSSProperties,
	item2: {
		flex: 1,
		minWidth: "100px",
	} as React.CSSProperties,
	item3: {
		flex: 1,
		minWidth: "125px",
	} as React.CSSProperties,
	item4: {
		flex: 1,
		minWidth: "270px",
	} as React.CSSProperties,
}

export default class ServiceUnit extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context)
		this.state = {
		}
	}

	render(): JSX.Element {
		if (this.props.hidden) {
			return <div></div>
		}

		let cards: JSX.Element[] = []
		for (let i = 0; i < 250; i++) {
			for (let deployment of this.props.unit.deployments) {
				cards.push(<Blueprint.Card compact={true} style={css.card}>
					<div className="layout horizontal wrap">
						<div className="layout center" style={css.checkBox}>
							<Blueprint.Checkbox
								style={css.check}
								checked={false}
								onChange={(): void => {

								}}
							/>
						</div>
						<div style={css.item1}>
							<PageInfo
								compact={true}
								style={css.info}
								fields={[
									{
										label: "Instance",
										value: deployment.instance_name,
									},
									{
										label: "Uptime",
										value: deployment.instance_uptime,
									},
								]}
							/>
						</div>
						<div style={css.item2}>
							<PageInfo
								compact={true}
								style={css.info}
								fields={[
									{
										label: "Status",
										value: deployment.instance_status,
									},
									{
										label: "Instance",
										value: deployment.instance_name,
									},
								]}
							/>
						</div>
						<div style={css.item3}>
							<PageInfo
								compact={true}
								style={css.info}
								fields={[
									{
										label: 'Public IPv4',
										value: deployment.public_ips,
									},
									{
										label: 'Private IPv4',
										value: deployment.private_ips,
									},
								]}
							/>
						</div>
						<div style={css.item4}>
							<PageInfo
								compact={true}
								style={css.info}
								fields={[
									{
										label: 'Public IPv6',
										value: deployment.public_ips6,
									},
									{
										label: 'Private IPv6',
										value: deployment.private_ips6,
									},
								]}
							/>
						</div>
					</div>
				</Blueprint.Card>)
			}
		}

		return <div className="layout horizontal wrap">
			<div style={css.box}>
				<Blueprint.CardList>
					{cards}
				</Blueprint.CardList>
			</div>
		</div>
	}
}
``