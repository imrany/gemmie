import React from "react";
import "./Chat.css";

const Chat = () => {
  return (
    <div className="chat_container">
      <div className="chat_title">
        <div className="chat_username">
          <h3>Gemmie</h3>
        </div>
        <div className="chat_header">
          <p>You are using Gemmie Demo, Google AI Hackathon</p>
        </div>
      </div>

      <div className="chat_body">
        <ion-icon name="chatbubble-outline"></ion-icon>
        <h1>Gemmie</h1>
        <p>
          Gemmie uses Gemini API to solve medical issues as a first aid
          assistant.
        </p>
      </div>
    </div>
  );
};

export default Chat;
