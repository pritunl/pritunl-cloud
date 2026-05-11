/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Blueprint from '@blueprintjs/core';
import * as Icons from '@blueprintjs/icons';
import * as MiscUtils from '../utils/MiscUtils';
import * as ImageTypes from '../types/ImageTypes';
import ImageDetailed from './ImageDetailed';
import CompletionStore from "../stores/CompletionStore";

interface Props {
	image: ImageTypes.ImageRo;
	selected: boolean;
	onSelect: (shift: boolean) => void;
	open: boolean;
	onOpen: () => void;
}

const css = {
	card: {
		display: 'table-row',
		width: '100%',
		padding: 0,
		boxShadow: 'none',
		cursor: 'pointer',
	} as React.CSSProperties,
	cardOpen: {
		display: 'table-row',
		width: '100%',
		padding: 0,
		boxShadow: 'none',
		position: 'relative',
	} as React.CSSProperties,
	select: {
		margin: '2px 0 0 0',
		paddingTop: '3px',
		minHeight: '18px',
	} as React.CSSProperties,
	name: {
		verticalAlign: 'top',
		display: 'table-cell',
		padding: '8px',
	} as React.CSSProperties,
	nameSpan: {
		margin: '1px 5px 0 0',
	} as React.CSSProperties,
	item: {
		verticalAlign: 'top',
		display: 'table-cell',
		padding: '9px',
		whiteSpace: 'nowrap',
	} as React.CSSProperties,
	icon: {
		marginRight: '3px',
	} as React.CSSProperties,
	bars: {
		verticalAlign: 'top',
		display: 'table-cell',
		padding: '8px',
		width: '30px',
	} as React.CSSProperties,
	bar: {
		height: '6px',
		marginBottom: '1px',
	} as React.CSSProperties,
	barLast: {
		height: '6px',
	} as React.CSSProperties,
	logo: {
		display: "inline-block",
		backgroundRepeat: "no-repeat",
		backgroundSize: "contain",
		backgroundPosition: "center",
		height: '19px',
		width: '19px',
		marginRight: '5px',
	} as React.CSSProperties,
};

export default class Image extends React.Component<Props, {}> {
	render(): JSX.Element {
		let image = this.props.image;

		if (this.props.open) {
			return <div
				className="bp5-card bp5-row"
				style={css.cardOpen}
			>
				<ImageDetailed
					image={this.props.image}
					selected={this.props.selected}
					onSelect={this.props.onSelect}
					onClose={(): void => {
						this.props.onOpen();
					}}
				/>
			</div>;
		}

		let cardStyle = {
			...css.card,
		};

		let orgClass = '';
		let orgIcon = '';
		let orgName = '';
		if (image.organization) {
			let org = CompletionStore.organization(image.organization);
			orgIcon = 'bp5-text-muted bp5-icon-people';
			orgName = org ? org.name : image.organization;
		} else {
			orgIcon = 'bp5-text-muted bp5-icon-globe';
			orgName = 'Public Image';
		}

		let name = image.name
		let icon: Blueprint.MaybeElement

		if (image.signed) {
			let imgSpl = image.key.split('_');
			let imgVer = imgSpl?.[1]?.split(".")[0] || ""
			let imgNameSpl = imgSpl[0].match(/^(.+?)(\d+)$/);
			let distro = imgNameSpl?.[1] || imgSpl[0]
			let version = imgNameSpl?.[2] || ""
			let matched = true

			switch (distro) {
				case "almalinux":
					name = `AlmaLinux ${version}`
					icon = <span style={css.logo} className="almalinux-logo"/>
					break
				case "alpinelinux":
					name = `Alpine Linux`
					icon = <span style={css.logo} className="alpinelinux-logo"/>
					break
				case "archlinux":
					name = `Arch Linux`
					icon = <span style={css.logo} className="archlinux-logo"/>
					break
				case "fedora":
					name = `Fedora ${version}`
					icon = <span style={css.logo} className="fedora-logo"/>
					break
				case "freebsd":
					name = `FreeBSD`
					icon = <span style={css.logo} className="freebsd-logo"/>
					break
				case "oraclelinux":
					name = `Oracle Linux ${version}`
					icon = <span style={css.logo} className="oraclelinux-logo"/>
					break
				case "rockylinux":
					name = `Rocky Linux ${version}`
					icon = <span style={css.logo} className="rockylinux-logo"/>
					break
				case "ubuntu":
					name = `Ubuntu ${version.slice(0, 2) + "." + version.slice(2)}`
					icon = <span style={css.logo} className="ubuntu-logo"/>
					break
				default:
					matched = false
			}

			if (matched && imgVer) {
				name += ` (${MiscUtils.parseImageDate(imgVer)})`
			}

			orgClass = 'bp5-text-intent-success';
			orgIcon = 'bp5-icon-endorsed';
			orgName = 'Signed Public Image';
		}

		let diskIcon = 'bp5-icon-box';
		switch (this.props.image.storage_class) {
			case 'aws_standard':
				diskIcon = 'bp5-icon-box';
				break;
			case 'aws_infrequent_access':
				diskIcon = 'bp5-icon-compressed';
				break;
			case 'aws_glacier':
				diskIcon = 'bp5-icon-snowflake';
				break;
			case 'oracle_standard':
				diskIcon = 'bp5-icon-box';
				break;
			case 'oracle_archive':
				diskIcon = 'bp5-icon-snowflake';
				break;
		}

		return <div
			className="bp5-card bp5-row"
			style={cardStyle}
			onClick={(evt): void => {
				let target = evt.target as HTMLElement;

				if (target.className.indexOf('open-ignore') !== -1) {
					return;
				}

				this.props.onOpen();
			}}
		>
			<div className="bp5-cell" style={css.name}>
				<div className="layout horizontal">
					<label
						className="bp5-control bp5-checkbox open-ignore"
						style={css.select}
					>
						<input
							type="checkbox"
							className="open-ignore"
							checked={this.props.selected}
							onChange={(evt): void => {
							}}
							onClick={(evt): void => {
								this.props.onSelect(evt.shiftKey);
							}}
						/>
						<span className="bp5-control-indicator open-ignore"/>
					</label>
					<div style={css.nameSpan}>
						{icon}{name}
					</div>
				</div>
			</div>
			<div className={'bp5-cell ' + orgClass} style={css.item}>
				<span
					style={css.icon}
					className={'bp5-icon-standard ' + orgIcon}
				/>
				{orgName}
			</div>
			<div className="bp5-cell" style={css.item}>
				<span
					style={css.icon}
					hidden={!image.key}
					className={'bp5-icon-standard bp5-text-muted ' + diskIcon}
				/>
				{image.key}
			</div>
		</div>;
	}
}
