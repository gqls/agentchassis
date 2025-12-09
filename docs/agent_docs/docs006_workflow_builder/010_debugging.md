1. Action Execution Flow
"Executing step in executeStep before executeLocalAction"
"just into executeLocalAction look for execCtx"
"Executing local action"
"Stored action result in storeActionResult"
"Stored action result under output_field"

2. HTML Generation Specific
   "Generating HTML content"
   "Processing HTML content"

3. LLM Calls & Responses
   "Executing LLM prompt action"
   "Rendered prompt template"
   "LLM response received"
   "result_preview"

4. Data Extraction & Context
   "Extracting data for AI agent"
   "Final template data"
   "in ExecuteLLMPromptAction Template Data"
   "in ExecuteLLMPromptAction data from which were trying to extract templatedata"

5. CollectedData Tracking
   "DEBUGaa: What have I done with CollectedData"
   "DEBUGaa: result_value"
   "DEBUGaa: executeLocalAction step"
   "DEBUGaa: executeLocalAction state after"

6. Response Handling
   "Stored response in output_field"
   "Created step data and stored response"
   "response_data"

Key fields to look for in the output:
kubectl logs <pod-name> | grep -B 5 -A 30 "generate_html"
    result_preview - first 200 chars of LLM response
    result_value - full result being stored
    Template Data - what's being sent to build the prompt
    CollectedData - what data is available at each step

Most Critical Search Commands:
# See the actual HTML being generated
kubectl logs <pod-name> | grep -A 20 "LLM response received"

# See what data is being passed to generate HTML
kubectl logs <pod-name> | grep -A 30 "Generating HTML content"

# See the prompt being sent to LLM
kubectl logs <pod-name> | grep -A 50 "Rendered prompt template"

# See what's in CollectedData at each step
kubectl logs <pod-name> | grep "DEBUGaa:"

# See action results being stored
kubectl logs <pod-name> | grep -A 10 "Stored action result"

# See the extracted template data
kubectl logs <pod-name> | grep -A 20 "Final template data"

The Golden Search (most likely to show the problem):
kubectl logs <pod-name> | grep -B 5 -A 30 "generate_html"

If extraction still fails after deployment:

Check logs for "=== MASTER EXTRACTOR START ===" message
