import { useLayoutEffect, useRef, useState } from "react";

import { BrowserOpenURL } from "../../../wailsjs/runtime";
import { ChatMessage } from "../../types/racetime";

type Props = {
  messages: ChatMessage[];
};

export default function ChatPanel({ messages }: Props) {
  const [activeChatTab, setActiveChatTab] = useState("main");
  const [unreadTabs, setUnreadTabs] = useState<Set<string>>(new Set());

  const chatRef = useRef<HTMLDivElement | null>(null);
  const wasAtBottomRef = useRef(true);

  const directMessageUsers = Array.from(
    new Map(
      messages
        .filter((msg) => msg.is_dm && msg.user)
        .map((msg) => [msg.user.id, msg.user]),
    ).values(),
  );

  const chatTabs = [
    {
      id: "main",
      label: "Main Chat",
    },
    ...directMessageUsers.map((user) => ({
      id: user.id,
      label: user.name,
    })),
  ];

  const filteredMessages = messages.filter((message) => {
    if (activeChatTab === "main") {
      return !message.is_dm;
    }

    return message.is_dm && message.user?.id === activeChatTab;
  });

  useLayoutEffect(() => {
    if (wasAtBottomRef.current && chatRef.current) {
      chatRef.current.scrollTop = chatRef.current.scrollHeight;
    }
  }, [messages]);

  useLayoutEffect(() => {
    const el = chatRef.current;

    if (!el) {
      return;
    }

    const onScroll = () => {
      wasAtBottomRef.current =
        el.scrollHeight - el.scrollTop - el.clientHeight < 50;
    };

    el.addEventListener("scroll", onScroll);

    return () => {
      el.removeEventListener("scroll", onScroll);
    };
  }, []);

  useLayoutEffect(() => {
    const nextUnread = new Set(unreadTabs);

    for (const message of messages) {
      if (message.is_dm && message.user && message.user.id !== activeChatTab) {
        nextUnread.add(message.user.id);
      }
    }

    setUnreadTabs(nextUnread);
  }, [messages, activeChatTab]);

  const formatChatTime = (timestamp: string) => {
    const date = new Date(timestamp);

    return date.toLocaleTimeString([], {
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const linkify = (text?: string | null) => {
    if (!text) {
      return text;
    }

    const urlRegex = /(https?:\/\/[^\s]+)/g;

    return text.split(urlRegex).map((part, index) => {
      if (part.match(/^https?:\/\//)) {
        return (
          <a
            key={index}
            href={part}
            onClick={(event) => {
              event.preventDefault();
              BrowserOpenURL(part);
            }}
          >
            {part}
          </a>
        );
      }

      return part;
    });
  };

  const renderMessage = (text?: string | null) => {
    if (!text) {
      return text;
    }

    // RaceTime highlights words with markers such as ##good##finished##.
    const parts = text.split(/(##(?:good|bad|bot)##.*?##)/g);

    return parts.map((part, index) => {
      const highlight = part.match(/^##(good|bad|bot)##(.*?)##$/);
      if (!highlight) {
        return <span key={index}>{linkify(part)}</span>;
      }

      return (
        <span
          key={index}
          className={
            highlight[1] === "good"
              ? "chatGood"
              : highlight[1] === "bad"
                ? "chatBad"
                : "chatBotHighlight"
          }
        >
          {highlight[2]}
        </span>
      );
    });
  };

  return (
    <div className="chatContainer">
      <div className="chatTabs">
        {chatTabs.map((tab) => (
          <button
            key={tab.id}
            className={activeChatTab === tab.id ? "chatTab active" : "chatTab"}
            onClick={() => {
              setActiveChatTab(tab.id);

              setUnreadTabs((prev) => {
                const next = new Set(prev);
                next.delete(tab.id);
                return next;
              });
            }}
          >
            {tab.label}
            {unreadTabs.has(tab.id) ? " •" : ""}
          </button>
        ))}
      </div>

      <div ref={chatRef} className="chatBox">
        {filteredMessages.map((message) => {
          const senderName = message.is_bot
            ? message.bot || "Bot"
            : (message.user?.name ?? "System");
          const isUserAction =
            message.is_system &&
            /\b(invites? .* to join|joins? the race|accepts? an invitation|is ready!|is no longer done|has .* from the race|has un-forfeited|has .*finished.* in .* place)\b/i.test(
              message.message,
            );
          const messageClass = message.is_bot
            ? "chatText chatBotMessage"
            : message.is_system
              ? isUserAction
                ? "chatText chatUserActionMessage"
                : "chatText chatSystemMessage"
              : "chatText chatUserMessage";

          return (
            <div
              key={message.id}
              className={message.is_dm ? "dmMessage" : "mainMessage"}
            >
              <div className={messageClass}>
                <span className="chatTimestamp">
                  {formatChatTime(message.posted_at)}
                </span>{" "}
                <span className="chatSender">
                  {senderName}
                  {!message.is_system && ":"}
                </span>
                {" "}
                {renderMessage(message.message)}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
