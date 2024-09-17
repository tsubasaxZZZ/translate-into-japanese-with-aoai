from openai import AzureOpenAI
from dotenv import load_dotenv
import os
from pydantic import BaseModel
import csv

load_dotenv()
client = AzureOpenAI(
  azure_endpoint = os.environ['BASE_OF_AZURE_ENDPOINT'], 
  api_key=os.environ['API_KEY'],  
  api_version="2024-08-01-preview"
)

# The model of the schema that is translated by OpenAI
class Translated(BaseModel):
    originalItem: str
    translatedItem: str
    originalDescription: str
    translatedDescription: str


# 1. Read from Book1.csv(Created by Excel) and it has two columns: Item and Description
data = []
with open('Book1.csv', 'r') as file:
    reader = csv.reader(file, dialect='excel')
    for row in reader:
        data.append(row)

# 2. Send data to OpenAI and translate it. Save the output to output.csv
for item, description in data:
    completion = client.beta.chat.completions.parse( 
        model=os.environ['DEPLOYMENT'],
        messages=[
            {"role": "system", "content": "You are an Azure expert. You are responsible for reviewing and translating user-provided Azure migration checklists. You should be aware that you are the expert and give the correct answers when responding to the user.You must use terminology in line with Azure and Microsoft documentation when translating. For example, 'autogrow' in SQL Server is '自動拡張'.' Not '自動成長'."},
            {"role": "user", "content": "Please translate into Japanese.Please make the Japanese more natural, even if the original meaning or expressions change slightly. Feel free to paraphrase."},
            {"role": "user", "content": f"Item(である調にしてください):{item}"},
            {"role": "user", "content": f"Description(ですます調にしてください):{description}"},
        ],
        response_format=Translated
    )
    print(completion.choices[0].message.parsed)
    with open('output.csv', 'a') as file:
        writer = csv.writer(file, dialect='excel')
        writer.writerow([completion.choices[0].message.parsed.originalItem, completion.choices[0].message.parsed.translatedItem, completion.choices[0].message.parsed.originalDescription, completion.choices[0].message.parsed.translatedDescription])
