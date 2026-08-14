import { useEffect, useState } from "react";
import "./App.css";

const API = "http://localhost:8080";

function App() {
  // =========================================================
  // AUTH
  // =========================================================

  const [isLoggedIn, setIsLoggedIn] = useState(
    localStorage.getItem("isLoggedIn") === "true"
  );

  const [authMode, setAuthMode] = useState("login");

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  const [authMessage, setAuthMessage] = useState("");

  // =========================================================
  // POLL
  // =========================================================

  const [question, setQuestion] = useState(
    "Which programming language do you like?"
  );

  const [options, setOptions] = useState([
    {
      id: "opt_1",
      text: "Python",
      votes: 0,
    },
    {
      id: "opt_2",
      text: "Java",
      votes: 0,
    },
  ]);

  const [pollId, setPollId] = useState("");

  // =========================================================
  // CREATE POLL
  // =========================================================

  const [newQuestion, setNewQuestion] = useState("");
  const [newOption1, setNewOption1] = useState("");
  const [newOption2, setNewOption2] = useState("");

  // =========================================================
  // UI
  // =========================================================

  const [voted, setVoted] = useState(false);
  const [message, setMessage] = useState("");

  // =========================================================
  // LOGIN
  // =========================================================

  const handleLogin = async (e) => {
    e.preventDefault();

    setAuthMessage("");

    if (!email || !password) {
      setAuthMessage("Please enter email and password.");
      return;
    }

    try {
      const response = await fetch(`${API}/api/login`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          email,
          password,
        }),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || "Login failed");
      }

      // Save login state
      localStorage.setItem("isLoggedIn", "true");
      localStorage.setItem("userEmail", email);

      setIsLoggedIn(true);

      setAuthMessage("");
      setEmail("");
      setPassword("");

      // Load latest poll after login
      loadLatestPoll();

    } catch (error) {
      console.log("Login error:", error);
      setAuthMessage(error.message);
    }
  };

  // =========================================================
  // SIGN UP
  // =========================================================

  const handleSignup = async (e) => {
    e.preventDefault();

    setAuthMessage("");

    if (!name || !email || !password) {
      setAuthMessage("Please fill all fields.");
      return;
    }

    try {
      const response = await fetch(`${API}/api/signup`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          username: name,
          email,
          password,
        }),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || "Signup failed");
      }

      setAuthMessage(
        "Account created successfully! Please login."
      );

      setName("");
      setEmail("");
      setPassword("");

      setAuthMode("login");

    } catch (error) {
      console.log("Signup error:", error);
      setAuthMessage(error.message);
    }
  };

  // =========================================================
  // LOGOUT
  // =========================================================

  const logout = () => {
    localStorage.removeItem("isLoggedIn");
    localStorage.removeItem("userEmail");

    setIsLoggedIn(false);

    setMessage("");
    setAuthMessage("");
  };

  // =========================================================
  // LOAD LATEST POLL
  // =========================================================

  const loadLatestPoll = async () => {
    try {
      const response = await fetch(`${API}/api/polls`);

      if (!response.ok) {
        return;
      }

      const data = await response.json();

      if (data.poll) {
        setQuestion(data.poll.question);
        setOptions(data.poll.options);
        setPollId(data.poll.id);
      }

    } catch (error) {
      console.log("Poll loading error:", error);
    }
  };

  // =========================================================
  // LOAD POLL AFTER LOGIN
  // =========================================================

  useEffect(() => {
    if (!isLoggedIn) {
      return;
    }

    loadLatestPoll();
  }, [isLoggedIn]);

  // =========================================================
  // IMPORTANT:
  // THIS MAKES TAB 1 AND TAB 2 SYNCHRONIZE
  //
  // Every 1 second both tabs ask backend for latest poll.
  // =========================================================

  useEffect(() => {
    if (!isLoggedIn) {
      return;
    }

    const interval = setInterval(() => {
      loadLatestPoll();
    }, 1000);

    return () => {
      clearInterval(interval);
    };
  }, [isLoggedIn]);

  // =========================================================
  // SYNC LOGIN BETWEEN TABS
  // =========================================================

  useEffect(() => {
    const syncLogin = (event) => {
      if (event.key === "isLoggedIn") {
        setIsLoggedIn(event.newValue === "true");
      }
    };

    window.addEventListener("storage", syncLogin);

    return () => {
      window.removeEventListener("storage", syncLogin);
    };
  }, []);

  // =========================================================
  // TOTAL VOTES
  // =========================================================

  const totalVotes = options.reduce(
    (total, option) => total + option.votes,
    0
  );

  // =========================================================
  // PERCENTAGE
  // =========================================================

  const percentage = (votes) => {
    if (totalVotes === 0) {
      return 0;
    }

    return Math.round((votes / totalVotes) * 100);
  };

  // =========================================================
  // CREATE POLL
  // =========================================================

  const createPoll = async (e) => {
    e.preventDefault();

    if (!newQuestion || !newOption1 || !newOption2) {
      setMessage("Please fill all fields.");
      return;
    }

    const pollData = {
      question: newQuestion,
      options: [
        {
          id: "opt_1",
          text: newOption1,
          votes: 0,
        },
        {
          id: "opt_2",
          text: newOption2,
          votes: 0,
        },
      ],
    };

    try {
      const response = await fetch(`${API}/api/polls`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(pollData),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(
          data.error || "Poll creation failed"
        );
      }

      // Update current tab immediately
      setQuestion(data.poll.question);
      setOptions(data.poll.options);
      setPollId(data.poll.id);

      // Reset form
      setNewQuestion("");
      setNewOption1("");
      setNewOption2("");

      setVoted(false);

      setMessage(
        "New poll created successfully! 🎉"
      );

      // Make sure latest poll is loaded
      loadLatestPoll();

    } catch (error) {
      console.log("Create poll error:", error);
      setMessage(error.message);
    }
  };

  // =========================================================
  // VOTE
  // =========================================================

  const vote = async (optionId) => {
    if (voted) {
      return;
    }

    if (!pollId) {
      setMessage("Please create a poll first.");
      return;
    }

    try {
      const response = await fetch(
        `${API}/api/polls/${pollId}/vote`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            optionId,
          }),
        }
      );

      const data = await response.json();

      if (!response.ok) {
        throw new Error(
          data.error || "Vote failed"
        );
      }

      // Immediate update in current tab
      setOptions((currentOptions) =>
        currentOptions.map((option) =>
          option.id === optionId
            ? {
                ...option,
                votes: option.votes + 1,
              }
            : option
        )
      );

      setVoted(true);

      setMessage(
        "Your vote has been recorded! 🎉"
      );

      // Get actual database value
      setTimeout(() => {
        loadLatestPoll();
      }, 300);

    } catch (error) {
      console.log("Vote error:", error);
      setMessage(error.message);
    }
  };

  // =========================================================
  // SHARE
  // =========================================================

  const sharePoll = async () => {
    try {
      await navigator.clipboard.writeText(
        window.location.href
      );

      setMessage(
        "Poll link copied successfully! 🔗"
      );

    } catch (error) {
      setMessage(
        "Copy the link from your browser."
      );
    }
  };

  // =========================================================
  // LOGIN / SIGNUP PAGE
  // =========================================================

  if (!isLoggedIn) {
    return (
      <div className="app">

        <header>
          <h1>📊 Live Polling App</h1>

          <p>
            Create polls. Vote. See live results.
          </p>
        </header>

        <main>

          <section className="poll-card create-card">

            <h2>
              {authMode === "login"
                ? "Sign In"
                : "Create Account"}
            </h2>

            {authMode === "signup" && (
              <input
                type="text"
                placeholder="Enter your name"
                value={name}
                onChange={(e) =>
                  setName(e.target.value)
                }
              />
            )}

            <input
              type="email"
              placeholder="Enter your email"
              value={email}
              onChange={(e) =>
                setEmail(e.target.value)
              }
            />

            <input
              type="password"
              placeholder="Enter your password"
              value={password}
              onChange={(e) =>
                setPassword(e.target.value)
              }
            />

            {authMode === "login" ? (
              <button
                className="create-button"
                onClick={handleLogin}
              >
                Sign In
              </button>
            ) : (
              <button
                className="create-button"
                onClick={handleSignup}
              >
                Sign Up
              </button>
            )}

            {authMessage && (
              <div className="message">
                {authMessage}
              </div>
            )}

            <p style={{ marginTop: "20px" }}>
              {authMode === "login"
                ? "Don't have an account?"
                : "Already have an account?"}
            </p>

            <button
              className="share-button"
              onClick={() => {
                setAuthMessage("");

                if (authMode === "login") {
                  setAuthMode("signup");
                } else {
                  setAuthMode("login");
                }
              }}
            >
              {authMode === "login"
                ? "Create an account"
                : "Back to Sign In"}
            </button>

          </section>

        </main>

      </div>
    );
  }

  // =========================================================
  // POLL PAGE
  // =========================================================

  return (
    <div className="app">

      {/* HEADER */}

      <header>

        <h1>
          📊 Live Polling App
        </h1>

        <p>
          Create polls. Vote. See live results.
        </p>

        <button
          className="share-button"
          onClick={logout}
        >
          Logout
        </button>

      </header>

      <main>

        {/* CREATE POLL */}

        <section className="poll-card create-card">

          <h2>
            Create a Poll
          </h2>

          <form onSubmit={createPoll}>

            <input
              type="text"
              placeholder="Enter your question"
              value={newQuestion}
              onChange={(e) =>
                setNewQuestion(e.target.value)
              }
            />

            <input
              type="text"
              placeholder="Option 1"
              value={newOption1}
              onChange={(e) =>
                setNewOption1(e.target.value)
              }
            />

            <input
              type="text"
              placeholder="Option 2"
              value={newOption2}
              onChange={(e) =>
                setNewOption2(e.target.value)
              }
            />

            <button
              className="create-button"
              type="submit"
            >
              + Create Poll
            </button>

          </form>

        </section>

        {/* LIVE POLL */}

        <section className="poll-card">

          <span className="live">
            ● LIVE POLL
          </span>

          <h2>
            {question}
          </h2>

          {/* OPTIONS */}

          <div className="options">

            {options.map((option) => (

              <button
                key={option.id}
                className="option"
                onClick={() =>
                  vote(option.id)
                }
                disabled={voted}
              >

                <span>
                  {option.text}
                </span>

                <strong>
                  {percentage(option.votes)}%
                </strong>

              </button>

            ))}

          </div>

          {/* MESSAGE */}

          {message && (
            <div className="message">
              {message}
            </div>
          )}

          {/* RESULTS */}

          <div className="results">

            <h3>
              Live Results
            </h3>

            {options.map((option) => (

              <div
                className="result"
                key={option.id}
              >

                <div className="result-info">

                  <span>
                    {option.text}
                  </span>

                  <span>
                    {option.votes} vote(s)
                  </span>

                </div>

                <div className="bar">

                  <div
                    className="bar-fill"
                    style={{
                      width: `${percentage(
                        option.votes
                      )}%`,
                    }}
                  ></div>

                </div>

              </div>

            ))}

            <p>
              Total votes: {totalVotes}
            </p>

          </div>

          {/* SHARE */}

          <button
            className="share-button"
            onClick={sharePoll}
          >
            🔗 Share Poll
          </button>

        </section>

      </main>

    </div>
  );
}

export default App;