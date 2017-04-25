// @flow

import React from 'react';

import Header from 'components/Header';
import Inner from 'components/Inner';

import styles from './styles.css';

const App = ({ children }: { children?: React.Children }) => (
  <div>
    <Header />

    <div className={styles.content}>
      <Inner>
        {children}
      </Inner>
    </div>
  </div>
);

export default App;
