import axios from 'axios';
import React from 'react';
import { render } from 'react-dom';
import { AppContainer } from 'react-hot-loader';
import { BrowserRouter } from 'react-router-dom';
import Redbox from 'redbox-react';

import Root from './root';
import { setupAxios, setupStore } from './setup';

function main() {
  let initialState = undefined;

  const reduxStateElement = document.getElementById('redux-state');
  if (reduxStateElement) {
    try {
      initialState = JSON.parse(reduxStateElement.innerText);
    } catch (e) {
      // do nothing
    }
  }

  const store = setupStore(initialState);
  setupAxios(axios, store);

  const reactRoot = document.getElementById('react-root');
  render(
    <AppContainer errorReporter={Redbox}>
      <BrowserRouter>
        <Root store={store} history={history} />
      </BrowserRouter>
    </AppContainer>,
    reactRoot
  );

  if (
    process.env.NODE_ENV !== 'production' && process.env.NODE_ENV !== 'test'
  ) {
    if (module.hot) {
      module.hot.accept('./root', () => {
        const NextApp = require('./root').default;

        render(
          <AppContainer errorReporter={Redbox}>
            <BrowserRouter>
              <NextApp />
            </BrowserRouter>
          </AppContainer>,
          reactRoot
        );
      });
    }
  }
}

if (document.readyState === 'complete') {
  main();
} else {
  document.addEventListener('DOMContentLoaded', main);
}
