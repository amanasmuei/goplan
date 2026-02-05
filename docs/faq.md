# GoPlan FAQ

## General Questions

### What is GoPlan?

GoPlan is an AI-powered task management platform that helps teams track work, estimate completion times, and gain insights into productivity patterns. It uses machine learning to predict task durations and identify potential bottlenecks.

### Who should use GoPlan?

GoPlan is designed for:
- Software development teams
- Project managers
- Individual contributors tracking their work
- Team leads managing workloads

### How is GoPlan different from other task managers?

GoPlan's key differentiators:
1. **AI-Powered Predictions**: Learns from historical data to predict task durations
2. **Acknowledgment Flow**: Ensures team members review and accept tasks
3. **Variance Analysis**: Tracks estimate accuracy to improve future planning
4. **Smart Insights**: Provides actionable productivity recommendations

---

## Tasks

### How do I create a task?

Click the **"+ New Task"** button in the header, or press `N` on your keyboard. Fill in the task details and click **"Create Task"**.

### What does "acknowledging" a task mean?

Acknowledging a task confirms that you've reviewed the task details and AI predictions. This step ensures:
- You understand what's expected
- You agree (or adjust) the time estimate
- The task is ready to be worked on

### Can I edit a task after creating it?

Yes! Click on any task to open it, then click the **Edit** button. You can modify the title, description, priority, due date, and other fields.

### How do I reassign a task?

1. Open the task
2. Click on the assignee field
3. Search for and select the new assignee
4. Save changes

### What happens to cancelled tasks?

Cancelled tasks are moved to a "Cancelled" status and excluded from active metrics. They remain in the system for historical reference but don't affect productivity calculations.

### Can I have sub-tasks?

Yes, GoPlan supports parent-child task relationships. When creating or editing a task, you can set a parent task to create a hierarchy.

---

## AI Predictions

### How does the AI predict task duration?

GoPlan's AI considers multiple factors:
- Historical data from similar tasks
- Your personal completion patterns
- Task complexity indicators (description length, keywords)
- Time of day and week patterns
- Current workload

### The AI prediction seems wrong. What should I do?

1. **Adjust the estimate** during acknowledgment - your input helps train the model
2. **Complete the task** and enter accurate actual time - this feedback improves predictions
3. Be patient - the AI learns over time and becomes more accurate with more data

### How accurate are the predictions?

Prediction accuracy improves as GoPlan learns from your team's data. Initial predictions may have 20-30% variance. With sufficient historical data (50+ completed tasks), accuracy typically improves to within 10-15%.

### Can I disable AI predictions?

While you can't disable predictions entirely, you can always override them during the acknowledgment step. The predictions are suggestions, not requirements.

---

## Analytics

### What metrics does GoPlan track?

- **Completion Rate**: Tasks finished on or before due date
- **Estimate Accuracy**: Variance between predicted and actual time
- **Velocity**: Tasks completed per time period
- **Time Distribution**: How time is spent across task types
- **Blocker Frequency**: How often and why tasks get blocked

### How is "productivity" calculated?

Productivity in GoPlan is based on:
- Tasks completed vs. planned
- Actual vs. estimated time (closer is better)
- Consistency over time

Note: GoPlan measures completion, not hours worked. Efficiency matters more than busyness.

### Can my manager see my analytics?

Team leads can see aggregated team metrics. Individual detailed analytics are private by default. Your organization can configure visibility settings.

### How far back does analytics data go?

GoPlan retains all historical data. Analytics dashboards typically show the last 30 days by default, but you can adjust the date range to view any period.

---

## Account & Settings

### How do I change my password?

GoPlan uses your company's Single Sign-On (SSO). Contact your IT department to reset your password.

### How do I update my notification preferences?

1. Click your profile icon in the top-right
2. Select **"Settings"**
3. Go to **"Notifications"**
4. Toggle individual notification types on/off

### Can I use GoPlan on mobile?

GoPlan is a web application optimized for desktop browsers. It works on mobile browsers but is best experienced on larger screens. A dedicated mobile app is planned for future releases.

### How do I change my timezone?

1. Go to **Settings**
2. Select **"Preferences"**
3. Update your timezone
4. Save changes

All dates and times will display in your selected timezone.

---

## Integrations

### Does GoPlan integrate with Slack?

Slack integration is planned for a future release. Currently, GoPlan sends notifications via in-app and email.

### Can I import tasks from other tools?

Import functionality is available for team administrators. Contact your admin to import data from:
- CSV files
- Jira
- Asana
- Trello

### Is there an API?

Yes, GoPlan has a REST API for integrations. API documentation is available at `/api/docs`. Contact your administrator for API access.

---

## Troubleshooting

### I can't log in

1. Verify you're using the correct URL
2. Clear your browser cache and cookies
3. Try a different browser
4. Contact your IT department if SSO is failing

### Tasks aren't syncing

1. Check your internet connection
2. Refresh the page (Ctrl+R or Cmd+R)
3. Log out and log back in
4. Contact support if the issue persists

### The page is loading slowly

1. Check your internet connection speed
2. Clear your browser cache
3. Disable browser extensions
4. Try using Chrome or Firefox for best performance

### I accidentally completed a task

Contact your team lead or admin. Tasks can be reopened with appropriate permissions.

### My data looks incorrect

1. Verify the date range in your filters
2. Check if any filters are excluding data
3. Refresh the page
4. Contact support if discrepancies persist

---

## Privacy & Security

### Is my data secure?

Yes. GoPlan uses:
- Industry-standard encryption (TLS 1.3)
- SOC 2 compliant infrastructure
- Regular security audits
- Role-based access control

### Who can see my tasks?

By default:
- You can see tasks assigned to you
- Team leads can see tasks for their team
- Admins can see all tasks

Your organization can customize these permissions.

### Can I export my data?

Yes, you can export your task data:
1. Go to **Settings**
2. Select **"Data Export"**
3. Choose the format (CSV, JSON)
4. Click **"Export"**

---

## Support

### How do I report a bug?

1. Click the feedback button in the bottom-right corner
2. Select **"Report Bug"**
3. Describe the issue and steps to reproduce
4. Submit

### How do I suggest a feature?

1. Click the feedback button
2. Select **"Feature Request"**
3. Describe your suggestion
4. Submit

### Who do I contact for help?

- **In-app help**: Click the `?` icon
- **Email**: goplan-support@yourcompany.com
- **Urgent issues**: Contact your team lead or IT department

---

*Last updated: January 2026*
