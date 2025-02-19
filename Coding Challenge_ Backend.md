# Coding Challenge: Wikipedia RecentChanges Discord Bot

Even though we're a gaming startup, real-time data handling is a critical part of our backend. We'd like to see how you manage live streams, filtering, and stats within a Discord context.

**Overview**  
You will build a Discord bot that consumes the [Wikipedia Recent Changes Stream](https://stream.wikimedia.org/v2/stream/recentchange). The bot will filter these changes by language and display the results upon user request. Bonus points if you include a feature to display daily statistics.

**Core Requirements**

1. **Data Ingestion:**  
     
   - Connect to the [Wikipedia Recent Changes Stream](https://stream.wikimedia.org/v2/stream/recentchange).  
   - Continuously listen for and ingest new change events.

   

2. **Language Filtering:**  
     
   - Provide a mechanism to filter events by a chosen language (e.g., `en`, `es`, etc.).  
   - Allow dynamic configuration, so that users in Discord can switch or set their preferred language.

   

3. **Discord Bot Commands:**  
     
   - **`!recent`**: Retrieves the most recent changes for the current or specified language.  
   - **`!setLang [language_code]`**: Sets a default language for the user/server session.  
   - Display relevant information about each change (title, URL, user, timestamp).

   

4. **Bonus Feature: Daily Stats**  
     
   - Track the number of changes per day for each language.  
   - Provide a command (e.g., `!stats [yyyy-mm-dd]`) to display how many changes occurred on that date for the chosen language.  
   - Consider storing data in a database or scalable data store (e.g., PostgreSQL, MongoDB, or Redis).

   

5. **Scalability (Extra Bonus)**  
     
   - Incorporate or outline how you would use technologies like Kafka, Spark, or similar frameworks to handle higher throughput.  
   - For instance, publish raw Wikipedia events to a Kafka topic, process them with Spark Streaming, and then consume the processed/filtered stream in your bot.  
   - In your README, discuss how you would scale this architecture for higher volumes of data.

   

6. **Project Structure & Documentation:**  
     
   - Include a README with setup instructions and a guide on how to run the bot.  
   - Explain design decisions and any trade-offs.  
   - Bonus for including tests or a CI/CD pipeline.

---

## Evaluation Criteria

- **Functionality:** Does your solution properly stream data, filter by language, and respond on Discord?  
- **Code Quality & Organization:** Is your code modular, testable, and well-structured?  
- **Documentation:** Can we easily follow setup instructions and understand your design choices?  
- **Creativity & Scalability:** Did you introduce unique features or consider scaling out your solution with robust data pipelines?

---

## How to Apply

1. **Resume/LinkedIn**: Send us a brief overview of your background, plus any relevant links (GitHub, portfolio, etc.).  
2. **Challenge Submission**: Upload your solution to a public repository (GitHub, GitLab) and share the link.  
3. **Short Write-Up**: In your README, describe your approach, tools used, and how you'd scale for large loads.

We look forward to seeing your creative solutions and potentially welcoming you to our growing gaming team\! Good luck and happy coding\!  

[Wikipedia SSE Stream] -> [Kafka] -> [Stats Service] -> [PostgreSQL]
                                 \-> [Discord Bot]
