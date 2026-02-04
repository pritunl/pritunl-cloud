/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as FirewallTypes from '../types/FirewallTypes';

interface Props {
	rule: FirewallTypes.Rule;
	onChange: (state: FirewallTypes.Rule) => void;
	onAdd: (prepend: boolean) => void;
	onRemove: () => void;
}

const css = {
	group: {
		width: '100%',
		maxWidth: '310px',
		marginTop: '5px',
	} as React.CSSProperties,
	sourceGroup: {
		width: '100%',
		maxWidth: '219px',
		marginTop: '5px',
	} as React.CSSProperties,
	protocol: {
		flex: '0 1 auto',
	} as React.CSSProperties,
	port: {
		width: '100%',
	} as React.CSSProperties,
	portBox: {
		flex: '1',
	} as React.CSSProperties,
	other: {
		flex: '0 1 auto',
		width: '52px',
		borderRadius: '0 3px 3px 0',
	} as React.CSSProperties,
};

export default class FirewallRule extends React.Component<Props, {}> {
	clone(): FirewallTypes.Rule {
		return {
			...this.props.rule,
		};
	}

	onAddSourceIp = (i: number, prepend: boolean): void => {
		let state = this.clone();

		let sourceIps = [
			...(state.source_ips || []),
		];

		sourceIps.splice(prepend ? i : i + 1, 0, '');
		state.source_ips = sourceIps;

		this.props.onChange(state);
	}

	onChangeSourceIp = (i: number, sourceIp: string): void => {
		let state = this.clone();

		let sourceIps = [
			...(state.source_ips || []),
		];

		sourceIps[i] = sourceIp;
		state.source_ips = sourceIps;

		this.props.onChange(state);
	}

	onRemoveSourceIp = (i: number): void => {
		let state = this.clone();

		let sourceIps = [
			...(state.source_ips || []),
		];

		sourceIps.splice(i, 1);
		state.source_ips = sourceIps;

		this.props.onChange(state);
	}

	render(): JSX.Element {
		let rule = this.props.rule;

		let port = rule.port;
		let placeholder = '';
		switch (rule.protocol) {
			case 'all':
				port = null;
				placeholder = 'Allow all traffic';
				break;
			case 'icmp':
				port = null;
				placeholder = 'Allow all ICMP traffic';
				break;
		}

		let sourceIps = (rule.source_ips || []);
		if (sourceIps.length === 0) {
			sourceIps.push('');
		}

		let sourceIpsDoms: JSX.Element[] = [];
		sourceIps.forEach((sourceIp: string, i: number): void => {
			sourceIpsDoms.push(
				<div
					className="bp5-control-group"
					style={css.sourceGroup}
					hidden={rule.protocol === "multicast" ||
						rule.protocol === "broadcast"}
					key={i}
				>
					<input
						className="bp5-input"
						style={css.port}
						type="text"
						autoCapitalize="off"
						spellCheck={false}
						placeholder="Source IP range"
						value={sourceIp}
						onChange={(evt): void => {
							this.onChangeSourceIp(i, evt.target.value);
						}}
					/>
					<button
						className="bp5-button bp5-minimal bp5-intent-danger bp5-icon-remove"
						onClick={(): void => {
							this.onRemoveSourceIp(i);
						}}
					/>
					<button
						className="bp5-button bp5-minimal bp5-intent-success bp5-icon-add"
						onClick={(evt): void => {
							this.onAddSourceIp(i, evt.shiftKey);
						}}
					/>
				</div>
			);
		});

		return <div>
			<div className="bp5-control-group" style={css.group}>
				<div className="bp5-select" style={css.protocol}>
					<select
						value={rule.protocol}
						onChange={(evt): void => {
							let state = this.clone();
							state.protocol = evt.target.value;

							if (state.protocol === 'all' || state.protocol === 'icmp') {
								state.port = null;
							}

							this.props.onChange(state);
						}}
					>
						<option value="all">All Traffic</option>
						<option value="icmp">ICMP</option>
						<option value="tcp">TCP</option>
						<option value="udp">UDP</option>
						<option value="multicast">Multicast</option>
						<option value="broadcast">Broadcast</option>
					</select>
				</div>
				<div style={css.portBox}>
					<input
						className="bp5-input"
						style={css.port}
						disabled={!!placeholder}
						type="text"
						autoCapitalize="off"
						spellCheck={false}
						placeholder={placeholder || 'Enter port range'}
						value={rule.port || ''}
						onChange={(evt): void => {
							let state = this.clone();
							state.port = evt.target.value;
							this.props.onChange(state);
						}}
					/>
				</div>
				<button
					className="bp5-button bp5-minimal bp5-intent-danger bp5-icon-remove"
					onClick={(): void => {
						this.props.onRemove();
					}}
				/>
				<button
					className="bp5-button bp5-minimal bp5-intent-success bp5-icon-add"
					onClick={(evt): void => {
						this.props.onAdd(evt.shiftKey);
					}}
				/>
			</div>
			{sourceIpsDoms}
		</div>;
	}
}
