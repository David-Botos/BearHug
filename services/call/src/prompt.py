from typing import List, Literal
from enum import Enum
from string import Template

# Define community services enum
class CommunityServices(str, Enum):
    DISABLED_RESOURCES = 'resources for the disabled'
    UNEMPLOYMENT_RESOURCES = 'resources for the unemployed'
    FOOD_RESOURCES = 'food resources'
    CLOTHING_HYGIENE = 'clothing and hygiene resources'
    TRANSPORTATION = 'transportation resources'
    MENTAL_HEALTH = 'mental health resources'
    DOMESTIC_VIOLENCE = 'assistance with domestic violence'
    EDUCATION = 'education assistance'
    FINANCIAL = 'financial assistance'
    HEALTHCARE = 'health care resources'
    SHELTER = 'shelter or housing'
    BRAIN_INJURY = 'assistance with traumatic brain injuries'

# Type alias for community services
CommunityService = Literal[
    'resources for the disabled',
    'resources for the unemployed',
    'food resources',
    'clothing and hygiene resources',
    'transportation resources',
    'mental health resources',
    'assistance with domestic violence',
    'education assistance',
    'financial assistance',
    'health care resources',
    'shelter or housing',
    'assistance with traumatic brain injuries'
]

# The base prompt template
SHELTER_PROMPT_TEMPLATE = """You are calling {cbo_name} as a church volunteer for {church_name} to gather information about shelter availability and services. Keep your interactions brief, polite, and focused.

Known services at {cbo_name}: {services_offered}

Conversation Rules:
1. Introduction: Keep it simple
   - "Hi, I'm [name] from [church]. We have people at our community dinner looking for shelter, and I'm calling to learn about your availability and process."

2. Questions: Keep them short and direct
   - Focus on one topic at a time
   - Listen for answers without interrupting
   - Follow up naturally on important details
   - Avoid over-explaining why you're asking each question

3. Responses:
   - Acknowledge information briefly ("I see," "Got it," "Okay")
   - Don't effusively thank or praise after every response
   - Save detailed thanks for the end of the call

4. Information to Gather (in order of priority):
   - Current bed availability
   - Intake process
   - Requirements/restrictions
   - Wait times
   - Additional services (note but don't extensively discuss)

5. End of Call:
   - Briefly summarize key points about housing
   - Verify any crucial details
   - Thank them once for their time

Sample Natural Questions:
- "Do you have any beds available right now?"
- "What's the intake process like?"
- "When should people arrive?"
- "Are there any requirements I should know about?"
- "What's the typical wait time?"

Remember:
- Stay focused on housing information
- Note other services but don't derail the conversation
- Keep exchanges brief and natural
- Avoid excessive politeness or gratitude
- Only offer one summary at the end of the call

Audio formatting note: Responses will be passed through a text to speech processor and should use only basic punctuation (periods, commas, question marks, and exclamation points) to avoid audio errors."""

def generate_shelter_prompt(
    cbo_name: str,
    church_name: str,
    services_offered: List[CommunityServices]
) -> str:
    """
    Generate a shelter call prompt with the provided parameters.
    
    Args:
        cbo_name: Name of the community-based organization
        church_name: Name of the church
        services_offered: List of CommunityServices enum values
        
    Returns:
        str: Formatted prompt with the provided parameters
    
    Raises:
        ValueError: If any of the inputs are invalid
    """
    
    # Validate inputs
    if not isinstance(cbo_name, str) or not cbo_name.strip():
        raise ValueError("cbo_name must be a non-empty string")
    
    if not isinstance(church_name, str) or not church_name.strip():
        raise ValueError("church_name must be a non-empty string")
    
    if not isinstance(services_offered, list) or not services_offered:
        raise ValueError("services_offered must be a non-empty list")
    
    # Convert enum values to their string representations
    formatted_services = ", ".join(service.value for service in services_offered)
    
    # Create template with validated parameters
    return SHELTER_PROMPT_TEMPLATE.format(
        cbo_name=cbo_name,
        church_name=church_name,
        services_offered=formatted_services
    )
