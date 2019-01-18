/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as MiscUtils from '../utils/MiscUtils';
import * as ImageTypes from '../types/ImageTypes';
import ImageDetailed from './ImageDetailed';
import OrganizationsStore from "../stores/OrganizationsStore";

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
		paddingTop: '1px',
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
};

export default class Image extends React.Component<Props, {}> {
	render(): JSX.Element {
		let image = this.props.image;

		if (this.props.open) {
			return <div
				className="pt-card pt-row"
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

		let active = true;

		let cardStyle = {
			...css.card,
		};
		if (!active) {
			cardStyle.opacity = 0.6;
		}

		let orgClass = '';
		let orgIcon = '';
		let orgName = '';
		if (image.organization) {
			let org = OrganizationsStore.organization(image.organization);
			orgIcon = 'pt-icon-people';
			orgName = org ? org.name : image.organization;
		} else {
			orgIcon = 'pt-icon-globe';
			orgName = 'Public Image';
		}

		if (image.signed) {
			orgClass = 'pt-text-intent-success';
			orgIcon = 'pt-icon-endorsed';
			orgName = 'Signed Public Image';
		}

		return <div
			className="pt-card pt-row"
			style={cardStyle}
			onClick={(evt): void => {
				let target = evt.target as HTMLElement;

				if (target.className.indexOf('open-ignore') !== -1) {
					return;
				}

				this.props.onOpen();
			}}
		>
			<div className="pt-cell" style={css.name}>
				<div className="layout horizontal">
					<label
						className="pt-control pt-checkbox open-ignore"
						style={css.select}
					>
						<input
							type="checkbox"
							className="open-ignore"
							checked={this.props.selected}
							onClick={(evt): void => {
								this.props.onSelect(evt.shiftKey);
							}}
						/>
						<span className="pt-control-indicator open-ignore"/>
					</label>
					<div style={css.nameSpan}>
						{image.name}
					</div>
				</div>
			</div>
			<div className={'pt-cell ' + orgClass} style={css.item}>
				<span
					style={css.icon}
					className={'pt-icon-standard pt-text-muted ' + orgIcon}
				/>
				{orgName}
			</div>
			<div className="pt-cell" style={css.item}>
				<span
					style={css.icon}
					hidden={!image.key}
					className="pt-icon-standard pt-text-muted pt-icon-compressed"
				/>
				{image.key}
			</div>
		</div>;
	}
}
