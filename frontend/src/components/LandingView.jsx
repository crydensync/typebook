import heroImage from "../assets/hero.png";

export default function LandingView({ onGetStarted, onLogin }) {
  return (
    <div className="landing">
      <div className="landing-hero">
        <img src={heroImage} alt="" className="landing-hero-img" />
        <h1>typebook</h1>
        <p className="landing-tagline">
          Quick notes, done right. Fast to capture, easy to find, backed by real
          account security — not an afterthought bolted on later.
        </p>
        <div className="landing-cta">
          <button className="btn btn-primary" onClick={onGetStarted}>
            Get started — it's free
          </button>
          <button className="btn btn-text" onClick={onLogin}>
            Log in
          </button>
        </div>
      </div>

      <div className="landing-features">
        <div className="landing-feature">
          <h3>Capture instantly</h3>
          <p>Title, body, done. No folders to set up before you can write anything down.</p>
        </div>
        <div className="landing-feature">
          <h3>Your sessions, your control</h3>
          <p>See every device logged into your account and revoke any one of them, anytime.</p>
        </div>
        <div className="landing-feature">
          <h3>Built on CrydenSync</h3>
          <p>
            Self-hosted, open authentication underneath — not a black-box login you have to
            trust blindly.
          </p>
        </div>
      </div>
    </div>
  );
}
