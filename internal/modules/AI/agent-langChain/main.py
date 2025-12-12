from typing import TypedDict, Annotated
import operator


class CustomerServiceState(TypedDict):
    # 这里的 Annotated[...] 是关键
    chat_history: Annotated[list, operator.add]
    refund: str
    intent: str

def classify_intent_node(state):
    return {"intent": "refund"}