// @flow

import React from 'react';

import styles from './styles.css';

const website = 'https://www.fknsrs.biz/';
const blog = 'https://www.fknsrs.biz/blog/don-statusnet-node-part-one-read-protocols.html';
const resume = 'https://www.fknsrs.biz/resume.html';

const App = ({ children }: { children?: React.Children }) => (
  <div>
    <header className={styles.header}>
      <a className={styles.title} href="/">DON</a>
    </header>

    <div className={styles.splash}>
      Hi! <a href={website}>I'm Conrad</a> and <a href={blog}>I made DON</a>.
      I'm also <a href={resume}>available for hire</a>!
    </div>

    <div className={styles.wrapper}>
      {children}
    </div>
  </div>
);

export default App;
