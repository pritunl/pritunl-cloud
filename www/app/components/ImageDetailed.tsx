/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as ImageTypes from '../types/ImageTypes';
import * as ImageActions from '../actions/ImageActions';
import * as MiscUtils from '../utils/MiscUtils';
import CompletionStore from "../stores/CompletionStore";
import PageInput from './PageInput';
import PageInfo from './PageInfo';
import PageSave from './PageSave';
import ConfirmButton from './ConfirmButton';
import PageTextArea from "./PageTextArea";

interface Props {
	image: ImageTypes.ImageRo;
	selected: boolean;
	onSelect: (shift: boolean) => void;
	onClose: () => void;
}

interface State {
	disabled: boolean;
	changed: boolean;
	message: string;
	image: ImageTypes.Image;
}

const css = {
	card: {
		position: 'relative',
		padding: '48px 10px 0 10px',
		width: '100%',
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
};

export default class ImageDetailed extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			changed: false,
			message: '',
			image: null,
		};
	}

	set(name: string, val: any): void {
		let image: any;

		if (this.state.changed) {
			image = {
				...this.state.image,
			};
		} else {
			image = {
				...this.props.image,
			};
		}

		image[name] = val;

		this.setState({
			...this.state,
			changed: true,
			image: image,
		});
	}

	onSave = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		ImageActions.commit(this.state.image).then((): void => {
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
						image: null,
						changed: false,
					});
				}
			}, 1000);

			setTimeout((): void => {
				if (!this.state.changed) {
					this.setState({
						...this.state,
						message: '',
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
		ImageActions.remove(this.props.image.id).then((): void => {
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
		let image: ImageTypes.Image = this.state.image ||
			this.props.image;

		let org = CompletionStore.organization(this.props.image.organization);
		let store = CompletionStore.storage(this.props.image.storage);

		let imgType = image.type;
		if (imgType) {
			imgType = imgType.charAt(0).toUpperCase() + imgType.slice(1);
		}

		let orgName = '';
		if (image.organization) {
			orgName = org ? org.name : image.organization;
		} else {
			orgName = 'Public Image';
		}

		if (image.signed) {
			orgName = 'Signed Public Image';
		}

		let storageClass = 'Default';
		switch (this.props.image.storage_class) {
			case 'aws_standard':
				storageClass = 'AWS Standard';
				break;
			case 'aws_infrequent_access':
				storageClass = 'AWS Standard-IA';
				break;
			case 'aws_glacier':
				storageClass = 'AWS Glacier';
				break;
			case 'oracle_standard':
				storageClass = 'Oracle Standard';
				break;
			case 'oracle_archive':
				storageClass = 'Oracle Archive';
				break;
		}

		return <td
			className="bp5-cell"
			colSpan={5}
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
							dialogLabel="Delete Image"
							confirmMsg="Permanently delete this image"
							confirmInput={true}
							items={[image.name]}
							disabled={this.props.image.type === 'public' ||
								this.state.disabled}
							onConfirm={this.onDelete}
						/>
					</div>
					<PageInput
						label="Name"
						help="Name of image"
						type="text"
						placeholder="Enter name"
						value={image.name}
						onChange={(val): void => {
							this.set('name', val);
						}}
					/>
					<PageTextArea
						label="Comment"
						help="Image comment."
						placeholder="Image comment"
						rows={3}
						value={image.comment}
						onChange={(val: string): void => {
							this.set('comment', val);
						}}
					/>
				</div>
				<div style={css.group}>
					<PageInfo
						fields={[
							{
								label: 'ID',
								value: this.props.image.id || 'Unknown',
							},
							{
								label: 'Storage',
								value: store ? store.name :
									this.props.image.storage || 'Unknown',
							},
							{
								label: 'Organization',
								value: orgName,
							},
							{
								label: 'Type',
								value: imgType || 'Unknown',
							},
							{
								label: 'Key',
								value: this.props.image.key || 'Unknown',
							},
							{
								label: 'Storage Class',
								value: storageClass,
							},
							{
								label: 'Last Modified',
								value: MiscUtils.formatDate(
									this.props.image.last_modified) || 'Unknown',
							},
							{
								label: 'ETag',
								value: this.props.image.etag || 'Unknown',
							},
						]}
					/>
				</div>
			</div>
			<PageSave
				style={css.save}
				hidden={!this.state.image && !this.state.message}
				message={this.state.message}
				changed={this.state.changed}
				disabled={this.state.disabled}
				light={true}
				onCancel={(): void => {
					this.setState({
						...this.state,
						changed: false,
						image: null,
					});
				}}
				onSave={this.onSave}
			/>
		</td>;
	}
}
