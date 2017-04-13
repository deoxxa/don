// @flow

import React from 'react';
import { Route } from 'react-router';
import { NavLink } from 'react-router-dom';

import Home from 'containers/Home';

import Flash from 'components/Flash';
import FontAwesome from 'components/FontAwesome';
import Header from 'components/Header';
import Inner from 'components/Inner';

import styles from './styles.css';

const website = 'https://www.fknsrs.biz/';
const blog = 'https://www.fknsrs.biz/blog/don-statusnet-node-part-one-read-protocols.html';
const resume = 'https://www.fknsrs.biz/resume.html';

const App = ({ children }: { children?: React.Children }) => (
  <div>
    <Header />

    <Flash success>
      Hi! <a href={website}>I'm Conrad</a>, and <a href={blog}>I made DON</a>.
      I'm also <a href={resume}>available for hire</a>!
    </Flash>

    <div className={styles.content}>
      <Inner>
        {children}
      </Inner>
    </div>
  </div>
);

export default App;
