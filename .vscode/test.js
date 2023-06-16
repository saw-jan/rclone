import fetch from 'node-fetch'
const url = 'http://localhost:8080/'

const h1 = {
  Authorization:
    'AWS4-HMAC-SHA256 Credential=BpYQ9SJl7TGxIKJY9ZlhGIOm7vXOCh1k7zRz5ZRd4x0ahqvNlfKQSF8WwwaMJ6T8/20130524/us-east-1/s3/aws4_request,SignedHeaders=host;range;x-amz-date,Signature=fe5f80f77d5fa3beca038a248ff027d0445342fe2855ddc963176630326f1024',
}

const h2 = {
  Authorization:
    'AWS4-HMAC-SHA256 Credential=egYUetyTqjcP6ZmgLHMo1QNsV3PrIFvp2NzUvAo6rmi4DCOamlJTbp1xcANQtjfP/20130524/us-east-1/s3/aws4_request,SignedHeaders=host;range;x-amz-date,Signature=fe5f80f77d5fa3beca038a248ff027d0445342fe2855ddc963176630326f1024',
}

const withTime = (timeout, header) => {
  return new Promise((resolve, reject) => {
    setTimeout(async () => {
      resolve(await fetch(url, { method: 'GET', ...header }))
    }, timeout)
  })
}

const requests = []

for (let i = 0; i < 5; i++) {
  // with time interval
  // requests.push(
  //   withTime(500 * i, {
  //     headers: h1,
  //   })
  // )
  // requests.push(
  //   withTime(1000 * i, {
  //     headers: h2,
  //   })
  // )
  //   all at once
  requests.push(
    fetch(url, {
      method: 'GET',
      headers: h1,
    })
  )
  requests.push(
    fetch(url, {
      method: 'GET',
      headers: h2,
    })
  )
  // one at a time
  // requests.push({
  //   headers: h1,
  // })
  // requests.push({
  //   headers: h2,
  // })
}

// async function fetchOneByOne() {
//   for (const header of requests) {
//     await fetch(url, { method: 'GET', ...header }).then(async (res) => {
//       const text = await res.text()
//       const match = text.match(/(?<=<Name>).*(?=<\/Name>)/)
//       console.log(match[0])
//     })
//   }
// }

// ;(async () => {
//   await fetchOneByOne()
// })()

Promise.all(requests).then(async (responses) => {
  let count = 1
  for (const res of responses) {
    const text = await res.text()
    const match = text.match(/(?<=<Name>).*(?=<\/Name>)/)
    console.log(match[0])
    console.log(count++)
  }
})
