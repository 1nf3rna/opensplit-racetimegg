export enum ConnectionStatus {
  Disconnected = 0,
  Connected = 1,
  Reconnecting = 2,
  WaitingForGame,
}

export type ConnectionState = {
  connection_status: ConnectionStatus;
  message: string;
};

export type RaceInfo = {
  UserID: string;
  Version: number;
  Goal: string;
  Game: string;
  RaceID: string;
  Info: string;

  StreamingRequired: boolean;

  EntrantCount: number;
  EntrantFinishedCount: number;
  EntrantInactiveCount: number;

  Entrants: Entrant[];

  Text: ChatMessage[];

  Ranked: boolean;
  AutoStart: boolean;

  Delay: number;

  Status: string;
  StatusVerbose: string;
  StatusHelpText: string;

  DisqualifyUnready: boolean;

  EndedAt: string | null;
  CancelledAt: string | null;
};

export type ChatMessage = {
  id: string;

  user: User;

  bot: string;

  direct_to: User;

  posted_at: string;

  message: string;
  message_plain: string;

  highlight: boolean;

  is_dm: boolean;
  is_bot: boolean;
  is_system: boolean;
  is_pinned: boolean;

  delay: string;
};

export type User = {
  id: string;

  full_name: string;
  name: string;

  discriminator: string;

  url: string;

  avatar: string;

  pronouns: string;

  flair: string;

  twitch_name: string;

  twitch_channel: string;

  can_moderate: boolean;
};

export type Entrant = {
  user: User;

  status: EntrantStatus;

  verbose_value: string;
  help_text: string;

  finish_time: string;
  finished_at: string;

  place: number;
  place_ordinal: string;

  score: number;
  score_change: number;

  comment: string;
  has_comment: boolean;

  stream_live: boolean;
  stream_override: boolean;
};

export type RaceActions = {
  canJoin: boolean;
  joinReason: string;
  joinAction: "join" | "accept_invite" | "request_invite";

  canLeave: boolean;
  leaveAction: "leave" | "decline_invite" | "cancel_invite";

  canReady: boolean;
  readyReason: string;

  canDone: boolean;
  doneReason: string;

  canForfeit: boolean;
  forfeitReason: string;

  streamBlocked: boolean;
};

export type EntrantStatus = {
  value: string;
  verbose_value: string;
  help_text: string;
};

export type UserStatus =
  | "requested"
  | "invited"
  | "declined"
  | "partitioned"
  | "not_joined"
  | "ready"
  | "not_ready"
  | "in_progress"
  | "done"
  | "dnf"
  | "dq";
