// @flow

import authentication from './authentication';
import type { State as AuthenticationState } from './authentication';
import publicTimeline from './publicTimeline';
import type { State as PublicTimelineState } from './publicTimeline';

export type State = {
  authentication: AuthenticationState,
  publicTimeline: PublicTimelineState,
};

export { authentication, publicTimeline };
