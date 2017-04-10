// @flow

import type axios from 'axios';
import { applyMiddleware, combineReducers, compose, createStore } from 'redux';
import { createLogger } from 'redux-logger';
import thunkMiddleware from 'redux-thunk';

import * as ducks from 'ducks';
import type { State } from 'ducks';

export function setupStore(initialState?: ?State): Object {
  let middleware = [thunkMiddleware];

  if (
    process.env.NODE_ENV !== 'production' && process.env.NODE_ENV !== 'test'
  ) {
    middleware = [...middleware, createLogger({ duration: true, diff: true })];
  }

  const createStoreWithMiddleware = compose(
    applyMiddleware(...middleware),
    typeof window !== 'undefined' &&
      typeof window.devToolsExtension === 'function'
      ? window.devToolsExtension()
      : s => s
  )(createStore);

  const store = createStoreWithMiddleware(combineReducers(ducks), initialState);

  if (
    process.env.NODE_ENV !== 'production' && process.env.NODE_ENV !== 'test'
  ) {
    if (module.hot) {
      module.hot.accept('./ducks', () => {
        store.replaceReducer(combineReducers(require('./ducks')));
      });
    }
  }

  return store;
}

export function setupAxios(axios: axios) {
  axios.interceptors.request.use(input => {
    let config = input;

    if (process.env.API_URL !== '') {
      config = { ...config, url: process.env.API_URL + config.url };
    }

    return config;
  });
}
