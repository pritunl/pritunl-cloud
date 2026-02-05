/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Blueprint from '@blueprintjs/core';
import * as Icons from '@blueprintjs/icons';
import * as ImageTypes from '../types/ImageTypes';
import * as MiscUtils from '../utils/MiscUtils';
import PageCustom from './PageCustom';

interface Props {
	images: ImageTypes.ImagesRo
	image: ImageTypes.Image;
	uefi: boolean
	showHidden: boolean
	disabled: boolean
	onChange: (img: ImageTypes.Image) => void
}

const css = {
	logo: {
		backgroundRepeat: "no-repeat",
    backgroundSize: "contain",
    backgroundPosition: "center",
		height: '19px',
		width: '19px',
	} as React.CSSProperties,
	imagesOpen: {
	} as React.CSSProperties,
	imagesMenu: {
		maxHeight: '460px',
		overflowY: "auto",
	} as React.CSSProperties,
};

export default class InstanceImages extends React.Component<Props, {}> {
	parseImage(img: ImageTypes.Image, button?: boolean): JSX.Element {
		let name = img.name
		let icon: Blueprint.IconName | Blueprint.MaybeElement = <Icons.Compressed/>

		if (img.signed) {
			let imgSpl = img.key.split('_');
			let imgVer = imgSpl?.[1]?.split(".")[0] || ""
			let imgNameSpl = imgSpl[0].match(/^(.+?)(\d+)$/);
			let distro = imgNameSpl?.[1] || imgSpl[0]
			let version = imgNameSpl?.[2] || ""
			let matched = true

			switch (distro) {
				case "almalinux":
					name = `AlmaLinux ${version}`
					icon = <div style={css.logo} className="almalinux-logo"/>
					break
				case "alpinelinux":
					name = `Alpine Linux`
					icon = <div style={css.logo} className="alpinelinux-logo"/>
					break
				case "archlinux":
					name = `Arch Linux`
					icon = <div style={css.logo} className="archlinux-logo"/>
					break
				case "fedora":
					name = `Fedora ${version}`
					icon = <div style={css.logo} className="fedora-logo"/>
					break
				case "freebsd":
					name = `FreeBSD`
					icon = <div style={css.logo} className="freebsd-logo"/>
					break
				case "oraclelinux":
					name = `Oracle Linux ${version}`
					icon = <div style={css.logo} className="oraclelinux-logo"/>
					break
				case "rockylinux":
					name = `Rocky Linux ${version}`
					icon = <div style={css.logo} className="rockylinux-logo"/>
					break
				case "ubuntu":
					name = `Ubuntu ${version.slice(0, 2) + "." + version.slice(2)}`
					icon = <div style={css.logo} className="ubuntu-logo"/>
					break
				default:
					matched = false
			}

			if (matched && imgVer) {
				name += ` (${MiscUtils.parseImageDate(imgVer)})`
			}
		}

		if (button) {
			return <Blueprint.Button
				alignText="left"
				icon={icon}
				rightIcon={<Icons.CaretDown/>}
				style={css.imagesOpen}
			>
				{name}
			</Blueprint.Button>
		}

		return <Blueprint.MenuItem
			key={img.id}
			roleStructure="listoption"
			icon={icon}
			onClick={(): void => {
				this.setState({
					...this.state,
					image: img,
				})
				this.props.onChange(img)
			}}
			text={name}
		/>
	}

	render(): JSX.Element {
		let hasImages = false
		let imagesSelect: JSX.Element[] = []
		let signedImages: ImageTypes.Image[] = []
		let otherImages: ImageTypes.Image[] = []
		let imagesVer = new Map<string, [number, ImageTypes.Image]>()
		let selectButton: JSX.Element

		if (this.props.images.length) {
			hasImages = true;
			for (let image of this.props.images) {
				if (this.props.uefi && image.firmware === 'bios') {
					continue;
				} else if (!this.props.uefi && image.firmware === 'uefi') {
					continue;
				}

				if (image.signed) {
					if (!this.props.showHidden) {
						let imgSpl = image.key.split('_');

						if (imgSpl.length >= 2 && imgSpl[imgSpl.length - 1].length >= 4) {
							let imgKey = imgSpl[0]
							let imgVer = parseInt(
								imgSpl[imgSpl.length - 1].substring(0, 4), 10);
							if (imgVer) {
								let curImg = imagesVer.get(imgKey);
								if (!curImg || imgVer > curImg[0]) {
									imagesVer.set(imgKey, [imgVer, image]);
								}
								continue
							}
						}
					} else {
						signedImages.push(image)
						continue
					}
				}

				otherImages.push(image)
			}

			const sortedVersionedImages = Array.from(imagesVer.entries())
				.sort((a, b) => MiscUtils.naturalSort(a[1][1].name, b[1][1].name));
			for (let [key, [ver, img]] of sortedVersionedImages) {
				imagesSelect.push(this.parseImage(img));
			}

			signedImages.sort((a, b) => MiscUtils.naturalSort(a.name, b.name));
			for (let img of signedImages) {
				imagesSelect.push(this.parseImage(img));
			}

			if (imagesSelect.length && otherImages.length) {
				imagesSelect.push(<Blueprint.MenuDivider
					key="menu-spec-divider"
				/>)
			}

			for (let img of otherImages) {
				imagesSelect.push(this.parseImage(img));
			}
		}

		if (this.props.image) {
			selectButton = this.parseImage(this.props.image, true)
		} else {
			selectButton = <Blueprint.Button
				alignText="left"
				icon={<Icons.Compressed/>}
				rightIcon={<Icons.CaretDown/>}
				text="Select Image"
				style={css.imagesOpen}
				disabled={this.props.disabled || !hasImages}
			/>
		}

		return <PageCustom
			label="Image"
			help="Starting image for instance."
		>
			<Blueprint.Popover
				content={<Blueprint.Menu style={css.imagesMenu}>
					{imagesSelect}
				</Blueprint.Menu>}
				placement="bottom"
			>
				{selectButton}
			</Blueprint.Popover>
		</PageCustom>
	}
}
