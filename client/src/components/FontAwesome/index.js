// @flow

import classNames from 'classnames/bind';
import styles from 'font-awesome/css/font-awesome.css';
import React, { Component, PropTypes } from 'react';

const cx = classNames.bind(styles);

export default class FontAwesome extends Component {
  props: { icon: string, className?: string };

  render() {
    const { icon, className, ...rest } = this.props;

    return <i {...rest} className={cx('fa', `fa-${icon}`, className)} />;
  }
}
