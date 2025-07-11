export interface CobaltData {
    status: string;
    url: string;
    filename: string;
}

export async function getCobalt(cobaltUrl: string, id: string): Promise<CobaltData> {
    const resp = await fetch(cobaltUrl, {
        method: 'POST',
        headers: {
            'accept': 'application/json',
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            url: 'https://youtube.com/watch?v='+id,
            audioFormat: 'best',
            downloadMode: 'audio',
            filenameStyle: 'basic',
            disableMetadata: false,
            alwaysProxy: false,
            localProcessing: false,
        })
    });

    return await resp.json() as CobaltData;
}