import React from "react";
import "./Sidebar.css";

const Sidebar = () => {
  return (
    <div className="sidebar_container">
      <div className="sidebar_title">
        <h3>Gemmie</h3>
        <ion-icon name="refresh-outline"></ion-icon>
      </div>

      <div className="sidebar_input">
        <input
          type="text"
          name="username"
          id="username"
          placeholder="Enter your username"
        />

        <button type="submit">Submit</button>
      </div>
    </div>
  );
};

export default Sidebar;
