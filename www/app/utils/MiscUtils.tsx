/// <reference path="../References.d.ts"/>
import React from "react";
import * as Blueprint from '@blueprintjs/core';
import * as ImageTypes from '../types/ImageTypes';
import * as Icons from '@blueprintjs/icons';

export class SyncInterval {
  private timer: number | null = null;
  private cancel: boolean = false;
  private readonly interval: number;
  private readonly action: () => Promise<any>;

  constructor(action: () => Promise<any>, interval: number) {
    this.action = action;
    this.interval = interval;
		this.start();
  }

  public start = async (): Promise<void> => {
    if (this.timer !== null) {
      clearTimeout(this.timer);
      this.timer = null;
    }

    this.cancel = false;

    const runSync = async (): Promise<void> => {
      if (this.cancel) return;

      try {
        await this.action();

        if (!this.cancel) {
          this.timer = window.setTimeout(() => {
            runSync();
          }, this.interval);
        }
      } catch (error) {
        console.error("Action error:", error);
        if (!this.cancel) {
          this.timer = window.setTimeout(() => {
            runSync();
          }, this.interval);
        }
      }
    };

    runSync();
  };

  public stop = (): void => {
    this.cancel = true;
    if (this.timer) {
      clearTimeout(this.timer);
      this.timer = null;
    }
  };
}

export function uuid(): string {
	return (+new Date() + Math.floor(Math.random() * 999999)).toString(36);
}

export function objectId(): string {
    const timestamp = Math.floor(Date.now() / 1000).toString(16);
    const randomBytes = Math.random().toString(16).substring(2, 12);
    const counter = Math.floor(Math.random() * 0xffffff).toString(16);
    return (timestamp + randomBytes + counter).padEnd(24, '0');
}

export function objectIdNil(objId: string): boolean {
	return !objId || objId == '000000000000000000000000';
}

export function zeroPad(num: number, width: number): string {
	if (num < Math.pow(10, width)) {
		return ('0'.repeat(width - 1) + num).slice(-width);
	}
	return num.toString();
}

export function capitalize(str: string): string {
	if (!str) {
		return str;
	}
	return str.charAt(0).toUpperCase() + str.slice(1);
}

export function titleCase(str: string): string {
	if (!str) {
		return str;
	}
	return str
		.toLowerCase()
		.split(' ')
		.map(word => word.charAt(0).toUpperCase() + word.slice(1))
		.join(' ');
}

export function formatAmount(amount: number): string {
	if (!amount) {
		return '-';
	}
	return '$' + (amount / 100).toFixed(2);
}

export function formatDate(dateStr: string): string {
	if (!dateStr || dateStr === '0001-01-01T00:00:00Z') {
		return '';
	}

	let date = new Date(dateStr);
	let str = '';

	let hours = date.getHours();
	let period = 'AM';

	if (hours > 12) {
		period = 'PM';
		hours -= 12;
	} else if (hours === 0) {
		hours = 12;
	}

	let day;
	switch (date.getDay()) {
		case 0:
			day = 'Sun';
			break;
		case 1:
			day = 'Mon';
			break;
		case 2:
			day = 'Tue';
			break;
		case 3:
			day = 'Wed';
			break;
		case 4:
			day = 'Thu';
			break;
		case 5:
			day = 'Fri';
			break;
		case 6:
			day = 'Sat';
			break;
	}

	let month;
	switch (date.getMonth()) {
		case 0:
			month = 'Jan';
			break;
		case 1:
			month = 'Feb';
			break;
		case 2:
			month = 'Mar';
			break;
		case 3:
			month = 'Apr';
			break;
		case 4:
			month = 'May';
			break;
		case 5:
			month = 'Jun';
			break;
		case 6:
			month = 'Jul';
			break;
		case 7:
			month = 'Aug';
			break;
		case 8:
			month = 'Sep';
			break;
		case 9:
			month = 'Oct';
			break;
		case 10:
			month = 'Nov';
			break;
		case 11:
			month = 'Dec';
			break;
	}

	str += day + ' ';
	str += date.getDate() + ' ';
	str += month + ' ';
	str += date.getFullYear() + ', ';
	str += hours + ':';
	str += zeroPad(date.getMinutes(), 2) + ':';
	str += zeroPad(date.getSeconds(), 2) + ' ';
	str += period;

	return str;
}

export function formatSinceLocal(dateStr: string): string {
	if (!dateStr || dateStr === "0001-01-01T00:00:00Z") {
		return "";
	}

	const now = new Date();
	let date = new Date(dateStr);
	date = new Date(date.getTime());
	const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);

	if (seconds < 60) {
		return `${seconds} seconds ago`;
	} else if (seconds < 3600) {
		const minutes = Math.floor(seconds / 60);
		return `${minutes} minutes ago`;
	} else if (seconds < 86400) {
		const hours = Math.floor(seconds / 3600);
		return `${hours} hours ago`;
	} else {
		const days = Math.floor(seconds / 86400);
		return `${days} days ago`;
	}
}

export function formatDateLocal(dateStr: string): string {
	if (!dateStr || dateStr === '0001-01-01T00:00:00Z') {
		return '';
	}

	let date = new Date(dateStr);
	date = new Date(date.getTime());
	let str = '';

	let hours = date.getHours();
	let period = 'AM';

	if (hours > 12) {
		period = 'PM';
		hours -= 12;
	} else if (hours === 0) {
		hours = 12;
	}

	let month;
	switch (date.getMonth()) {
		case 0:
			month = 'Jan';
			break;
		case 1:
			month = 'Feb';
			break;
		case 2:
			month = 'Mar';
			break;
		case 3:
			month = 'Apr';
			break;
		case 4:
			month = 'May';
			break;
		case 5:
			month = 'Jun';
			break;
		case 6:
			month = 'Jul';
			break;
		case 7:
			month = 'Aug';
			break;
		case 8:
			month = 'Sep';
			break;
		case 9:
			month = 'Oct';
			break;
		case 10:
			month = 'Nov';
			break;
		case 11:
			month = 'Dec';
			break;
	}

	str += month + ' ';
	str += zeroPad(date.getDate(), 2) + ', ';
	str += date.getFullYear() + ' ';
	str += zeroPad(hours, 2) + ':';
	str += zeroPad(date.getMinutes(), 2) + ':';
	str += zeroPad(date.getSeconds(), 2) + ' ';
	str += period;

	return str;
}

export function formatDateShort(dateStr: string): string {
	if (!dateStr || dateStr === '0001-01-01T00:00:00Z') {
		return '';
	}

	let date = new Date(dateStr);
	let curDate = new Date();

	let month;
	switch (date.getMonth()) {
		case 0:
			month = 'Jan';
			break;
		case 1:
			month = 'Feb';
			break;
		case 2:
			month = 'Mar';
			break;
		case 3:
			month = 'Apr';
			break;
		case 4:
			month = 'May';
			break;
		case 5:
			month = 'Jun';
			break;
		case 6:
			month = 'Jul';
			break;
		case 7:
			month = 'Aug';
			break;
		case 8:
			month = 'Sep';
			break;
		case 9:
			month = 'Oct';
			break;
		case 10:
			month = 'Nov';
			break;
		case 11:
			month = 'Dec';
			break;
	}

	let str = month + ' ' + date.getDate();

	if (date.getFullYear() !== curDate.getFullYear()) {
		str += ' ' + date.getFullYear();
	}

	return str;
}

export function formatDateShortTime(dateStr: string): string {
	if (!dateStr || dateStr === '0001-01-01T00:00:00Z') {
		return '';
	}

	let date = new Date(dateStr);
	let curDate = new Date();

	let month;
	switch (date.getMonth()) {
		case 0:
			month = 'Jan';
			break;
		case 1:
			month = 'Feb';
			break;
		case 2:
			month = 'Mar';
			break;
		case 3:
			month = 'Apr';
			break;
		case 4:
			month = 'May';
			break;
		case 5:
			month = 'Jun';
			break;
		case 6:
			month = 'Jul';
			break;
		case 7:
			month = 'Aug';
			break;
		case 8:
			month = 'Sep';
			break;
		case 9:
			month = 'Oct';
			break;
		case 10:
			month = 'Nov';
			break;
		case 11:
			month = 'Dec';
			break;
	}

	let str = month + ' ' + date.getDate();

	if (date.getFullYear() !== curDate.getFullYear()) {
		str += ' ' + date.getFullYear();
	} else if (date.getMonth() === curDate.getMonth() &&
			date.getDate() === curDate.getDate()) {
		let hours = date.getHours();
		let period = 'AM';

		if (hours > 12) {
			period = 'PM';
			hours -= 12;
		} else if (hours === 0) {
			hours = 12;
		}

		str = hours + ':';
		str += zeroPad(date.getMinutes(), 2) + ':';
		str += zeroPad(date.getSeconds(), 2) + ' ';
		str += period;
	}

	return str;
}

export function humanReadableSpeedMb(speedMb: number): string {
  if (!speedMb || speedMb <= 0) {
    return '';
  }

  if (speedMb >= 1000) {
    return `${(speedMb / 1000).toFixed(1)} GB/s`;
  } else {
    return `${speedMb.toFixed(1)} MB/s`;
  }
}

export function highlightMatch(input: string, query: string): React.ReactNode {
	if (!query) {
		return input;
	}

	let index = input.toLowerCase().indexOf(query.toLowerCase())
	if (index === -1) {
		return input;
	}

	return <span>
		{input.substring(0, index)}
		<b>{input.substring(index, index + query.length)}</b>
		{input.substring(index + query.length)}
	</span>;
}

export function parseImageDate(dateString: string): string {
  if (!dateString || typeof dateString !== 'string') {
		return dateString
  }

  const cleanedDate = dateString.replace(/\D/g, '');

  if (cleanedDate.length === 6) {
    const year = cleanedDate.substring(0, 2);
    const month = cleanedDate.substring(2, 4);

    const monthNum = parseInt(month, 10);
    if (monthNum < 1 || monthNum > 12) {
			return dateString
    }

    return `${month}/${year}`;
  } else if (cleanedDate.length === 4) {
    const year = cleanedDate.substring(0, 2);
    const month = cleanedDate.substring(2, 4);

    const monthNum = parseInt(month, 10);
    if (monthNum < 1 || monthNum > 12) {
			return dateString
    }

    return `${month}/${year}`;
  }

	return dateString
}
