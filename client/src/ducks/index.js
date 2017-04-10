// @flow

import publicTimeline from './publicTimeline';
import type { State as PublicTimelineState } from './publicTimeline';

export type State = {
  publicTimeline: PublicTimelineState,
};

export { publicTimeline };
