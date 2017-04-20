// @flow

import axios from 'axios';

export type ASPerson = {
  id: string,
  host: string,
  firstSeen: string,
  permalink: string,
  displayName: ?string,
  avatar: ?string,
  summary: ?string,
};

export type ASObject = {
  id: string,
  name: ?string,
  summary: ?string,
  representativeImage: ?string,
  permalink: ?string,
  objectType: ?string,
  content: ?string,
};

export type ASActivity = {
  id: string,
  permalink: string,
  actorID: ?string,
  actor: ?ASPerson,
  objectID: string,
  object: ASObject,
  verb: string,
  time: string,
  title: ?string,
  inReplyToID: ?string,
  inReplyToURL: ?string,
};

export type State = {
  loading: boolean,
  activities: ?Array<ASActivity>,
  error: ?Error,
};

export const publicTimelineError = (error: Error) => ({
  type: 'don/publicTimeline/ERROR',
  payload: { error },
});
export const publicTimelineLoaded = (activities: Array<ASActivity>) => ({
  type: 'don/publicTimeline/LOADED',
  payload: { activities },
});
export const publicTimelineLoading = () => ({
  type: 'don/publicTimeline/LOADING',
  payload: {},
});

export const publicTimelineFetch = () => (dispatch: (a: Object) => void) => {
  dispatch(publicTimelineLoading());

  return axios
    .get('/')
    .then(
      ({ data: { publicTimeline: { activities } } }) =>
        dispatch(publicTimelineLoaded(activities)),
      error => dispatch(publicTimelineError(error))
    );
};

const defaultState = {
  loading: false,
  activities: null,
  error: null,
};

export default (
  state: State = defaultState,
  action:
    | { type: 'don/publicTimeline/ERROR', payload: { error: Error } }
    | {
        type: 'don/publicTimeline/LOADED',
        payload: { activities: Array<ASActivity> },
      }
    | { type: 'don/publicTimeline/LOADING', payload: {} }
) => {
  switch (action.type) {
    case 'don/publicTimeline/ERROR':
      return {
        ...state,
        loading: false,
        error: action.payload.error,
      };
    case 'don/publicTimeline/LOADED':
      return {
        ...state,
        loading: false,
        error: null,
        activities: action.payload.activities,
      };
    case 'don/publicTimeline/LOADING':
      return {
        ...state,
        loading: true,
        error: null,
      };
    default:
      return state;
  }
};
